package amp

import (
  "errors"
  "fmt"
  "io"
  "net/rpc"
  "strconv"
  "sync"
)

type serverCodec struct {
  conn io.ReadWriteCloser
  req  Box

  mutex   sync.Mutex // protects seq, pending
  seq     uint64
  pending map[uint64]*uint64
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
  return &serverCodec{
    conn:    conn,
    pending: make(map[uint64]*uint64),
  }
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
  box, err := Decode(c.conn)
  if err != nil {
    return err
  }
  c.req = box
  r.ServiceMethod = fmt.Sprintf("AMP.%v", CamelCase(box["_command"]))
  var id *uint64
  if _, ok := box["_ask"]; ok {
    i, err := strconv.Atoi(box["_ask"])
    if err != nil {
      return err
    }
    temp := uint64(i)
    id = &temp
  } else {
    id = nil
  }
  c.mutex.Lock()
  c.seq++
  c.pending[c.seq] = id
  r.Seq = c.seq
  c.mutex.Unlock()
  return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
  return c.req.Unmarshal(x)
}

var errorRegistry = make(map[string]string)

func RegisterErrorCode(code string, msg string) {
  errorRegistry[code] = msg
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
  c.mutex.Lock()
  id, ok := c.pending[r.Seq]
  if !ok {
    c.mutex.Unlock()
    return errors.New("invalid sequence number in response")
  }
  delete(c.pending, r.Seq)
  c.mutex.Unlock()

  if id == nil {
    return nil // No response needed.
  }
  if r.Error != "" {
    var resp Box
    if msg, ok := errorRegistry[r.Error]; ok {
      resp = Box{"_error": fmt.Sprintf("%d", id), "_error_code": r.Error, "_error_description": msg}
    } else {
      resp = Box{"_error": fmt.Sprintf("%d", id), "_error_code": "UNKNOWN", "_error_description": "Unknown Error"}
    }
    repr, err := resp.Serialize()
    if err != nil {
      return fmt.Errorf("unable to serialize error response: %v", err)
    }
    _, err = c.conn.Write(repr)
    if err != nil {
      return fmt.Errorf("unable to write to the connection: %v", err)
    }
    return nil
  }
  resp, ok := x.(Box)
  if !ok {
    return fmt.Errorf("return value is not a Box, got: %+v", x)
  }
  resp["_answer"] = fmt.Sprintf("%d", &id)
  repr, err := resp.Serialize()
  if err != nil {
    return fmt.Errorf("unable to serialize response: %v", err)
  }
  _, err = c.conn.Write(repr)
  if err != nil {
    return fmt.Errorf("unable to write to the connection: %v", err)
  }
  return nil
}

func (c *serverCodec) Close() error {
  return c.conn.Close()
}

func ServeConn(conn io.ReadWriteCloser) {
  rpc.ServeCodec(NewServerCodec(conn))
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
  return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
  return '0' <= c && c <= '9'
}

// CamelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
func CamelCase(s string) string {
  if s == "" {
    return ""
  }
  t := make([]byte, 0, 32)
  i := 0
  if s[0] == '_' {
    // Need a capital letter; drop the '_'.
    t = append(t, 'X')
    i++
  }
  // Invariant: if the next letter is lower case, it must be converted
  // to upper case.
  // That is, we process a word at a time, where words are marked by _ or
  // upper case letter. Digits are treated as words.
  for ; i < len(s); i++ {
    c := s[i]
    if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
      continue // Skip the underscore in s.
    }
    if isASCIIDigit(c) {
      t = append(t, c)
      continue
    }
    // Assume we have a letter now - if not, it's a bogus identifier.
    // The next word is a sequence of characters that must start upper case.
    if isASCIILower(c) {
      c ^= ' ' // Make it a capital letter.
    }
    t = append(t, c) // Guaranteed not lower case.
    // Accept lower case sequence that follows.
    for i+1 < len(s) && isASCIILower(s[i+1]) {
      i++
      t = append(t, s[i])
    }
  }
  return string(t)
}
