package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Writer struct {
	writer *bufio.Writer
}
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer : bufio.NewWriter(w),
	}
}

func (w *Writer) Write(v Value) error {
	var err error

	switch v.Type{
	case TypeSimpleString:
		_,err = w.writer.WriteString("+" + v.Str + "\r\n")
	case TypeInteger:
		_,err = w.writer.WriteString(":" + strconv.FormatInt(v.Num, 10) + "\r\n")
	case TypeError:
		_,err = w.writer.WriteString("+" + v.Str + "\r\n")
	case TypeBulkString:
		if v.IsNull {
			_,err = w.writer.WriteString("$-1\r\n")
		}else{
			header := fmt.Sprintf("$%d\r\n",len(v.Bulk))
			if _,err = w.writer.WriteString(header); err != nil {
				return err
			}
			if _,err = w.writer.WriteString(string(v.Bulk)); err != nil{
				return err
			}
			_,err = w.writer.WriteString("\r\n")
		}
	case TypeArray:
		if v.IsNull{
			_,err = w.writer.WriteString("*-1\r\n")
		}else{
			header := fmt.Sprintf("*%d\r\n",len(v.Array))
			if _,err = w.writer.WriteString(header); err != nil{
				return err
			}
			for _,i := range v.Array{
				err := w.Write(i)
				if err != nil {
					return err
				}
			}
		}
	default:
		return ErrUnknownType
	}
	if err != nil{
		return err
	}
	return w.writer.Flush()
}
