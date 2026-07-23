package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrInvalidSyntax = errors.New("invalid RESP syntax")
	ErrUnknownType   = errors.New("unknown RESP data type")
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		reader:bufio.NewReader(rd),
	}
}

func (r *Reader) Read() (Value, error) {
	dataType,err := r.reader.ReadByte()

	if err != nil {
		return Value{},err
	}

	switch dataType{
	case TypeSimpleString:
		return r.readSimpleString()
	case TypeArray:
		return r.readArray()
	case TypeError:
		return r.readError()
	case TypeInteger:
		return r.readInteger()
	case TypeBulkString:
		return r.readBulkString()
	default:
		return Value{}, fmt.Errorf("%w: %c", ErrUnknownType, dataType)
	}

} 

func (r *Reader) readLine() ([]byte,error) {
	line,err := r.reader.ReadBytes('\n')
	if err != nil{
		return nil,err
	}
	if len(line) < 2 || line[len(line) - 2] != '\r'{
		return nil,ErrInvalidSyntax
	}

	return line[:len(line) - 2],nil
}

func (r *Reader) readIntLine() (int64,error){
	line,err := r.readLine()
	
	if err != nil{
		return 0,err
	}
	num,err := strconv.ParseInt(string(line),10,64)
	return num,err

}

func (r *Reader) readSimpleString() (Value,error) {
	line,err := r.readLine()
	if err != nil{
		return Value{},err
	}
	return Value{Type: TypeSimpleString,Str: string(line)},nil
}

func (r *Reader) readBulkString() (Value, error){
	lenght,err := r.readIntLine()
	if err != nil{
		return Value{},err
	}

	if lenght == -1 {
		return Value{Type: TypeBulkString,IsNull: true},nil
	}

	if lenght < 0{
		return Value{},ErrInvalidSyntax
	}

	buf := make([]byte,lenght)
	if _,err := io.ReadFull(r.reader,buf); err != nil{
		return Value{},err
	}
	check := make([]byte,2)
	if _,err := io.ReadFull(r.reader,check); err != nil{
		return Value{},err
	}
	if check[0] != '\r' || check[1] != '\n'{
		return Value{},ErrInvalidSyntax
	}

	return Value{Type: TypeBulkString,Bulk: buf},nil

}

func (r *Reader) readArray() (Value,error){
	cnt,err := r.readIntLine()

	if err != nil{
		return Value{},err
	}

	if cnt == -1{
		return Value{Type: TypeArray,IsNull: true},nil
	}
	if cnt < 0{
		return Value{},ErrInvalidSyntax
	}

	items := make([]Value,cnt)

	for i:=int64(0);i<cnt;i++{
		val,err := r.Read()
		if err != nil{
			return Value{},err
		}
		items[i] = val
	}
	return Value{Type: TypeArray,Array: items},nil
}

func (r *Reader) readError() (Value,error){
	line,err := r.readLine()
	
	if err != nil {
		return Value{},err
	}

	return Value{Type: TypeError, Str: string(line)},nil
}

func (r *Reader) readInteger() (Value, error){
	num,err := r.readIntLine()

	if err != nil {
		return Value{},err
	}

	return Value{Type: TypeInteger,Num: num},nil
}
