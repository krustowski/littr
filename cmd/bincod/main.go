package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type Example struct {
	One   int64
	Two   string
	Three float64
}

func (e Example) MarshalBinary() []byte {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, e.One, e.Two, e.Three)

	return buf.Bytes()
}

func (e *Example) UnmarshalBinary(data *[]byte) error {
	buf := bytes.NewBuffer(*data)

	_, err := fmt.Fscanln(buf, e.One, e.Two, e.Three)

	return err
}

func main() {
	var (
		ex = []Example{
			{
				One:   99785521455,
				Two:   "test string that is a bit long than you would normally, casually expected in such scenarios",
				Three: 3.1415,
			},
			{
				One:   778,
				Two:   "ok stop",
				Three: 999.99999,
			},
		}
		wbuf bytes.Buffer
	)

	enc := gob.NewEncoder(&wbuf)
	if err := enc.Encode(ex); err != nil {
		log.Fatal("encode: ", err)
	}

	os.WriteFile("example.bin", wbuf.Bytes(), 0600)

	//
	//
	//

	rb, err := os.ReadFile("example.bin")
	if err != nil {
		log.Fatal("read: ", err)
	}

	rbuf := bytes.NewReader(rb)

	dec := gob.NewDecoder(rbuf)

	var dex []Example
	if err := dec.Decode(&dex); err != nil {
		log.Fatal("decode: ", err)
	}

	fmt.Println(dex)
}
