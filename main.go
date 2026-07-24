package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/NiravShah1729/semdb/protocol"
	"github.com/NiravShah1729/semdb/store"
)

func main() {
	kvStore := store.NewStore()

	listener,err := net.Listen("tcp",":8080")
	if err != nil{
		log.Fatalf("Failed to bind to port 8080: %v",err)
	}
	defer listener.Close()

	log.Println("Server listening on 8080")

	for {
		conn,err := listener.Accept()

		if err != nil{
			log.Printf("Error accepting connection: %v",err)
			continue
		}
		go handleConnection(conn,kvStore)
	}
}

func handleConnection(conn net.Conn, kv *store.Store){
	defer conn.Close()

	writer := protocol.NewWriter(conn)
	reader := protocol.NewReader(conn)

	for {
		val,err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Printf("Connection read error:%v",err)
			return
		}

		if val.Type != protocol.TypeArray {
			writer.Write(protocol.Value{
				Type: protocol.TypeError,
				Str: "Err command must be a RESP array",
			})
			continue
		}

		if len(val.Array) == 0{
			writer.Write(protocol.Value{
				Type: protocol.TypeArray,
				Str: "Err empty command",
			})
			continue
		}

		cmd := strings.ToUpper(val.Array[0].String())
		args := val.Array[1:]

		switch cmd {
		case "PING":
			if len(args) > 0 {
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(args[0].String()),
				})
			}else{
				writer.Write(protocol.Value{
					Type :protocol.TypeSimpleString,
					Str: "PONG",
 				})
			}
		case "SET":
			if len(args) < 2{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments for set command",
				})
				continue
			}
			key := args[0].String()
			val := args[1].String()
			kv.Set(key,val)
			writer.Write(protocol.Value{
				Type: protocol.TypeSimpleString,
				Str: "OK",
			})
		case "GET":
			if len(args) < 1{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments for get command",
				})
				continue
			}
			key := args[0].String()
			res,ok := kv.Get(key)
			if !ok {
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					IsNull: true,
				})
			}else{
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(res),
				})
			}
		case "EXISTS":
			if len(args) < 1 {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments for exists command",
				})
				continue
			}
			keys := make([]string,len(args))
			for i,arg := range args {
				keys[i] = arg.String()
			}

			count := kv.Exists(keys...)
			writer.Write(protocol.Value{
				Type: protocol.TypeInteger,
				Num: int64(count),
			})
		case "DEL":
			if len(args) < 1 {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments for delete command",
				})
				continue
			}
			keys := make([]string,len(args))
			for i,arg := range args {
				keys[i] = arg.String()
			}

			count := kv.Del(keys...)
			writer.Write(protocol.Value{
				Type: protocol.TypeInteger,
				Num: int64(count),
			})
		default:
			writer.Write(protocol.Value{
				Type: protocol.TypeArray,
				Str: fmt.Sprintf("ERR unknown command %s",cmd),
			})
		}
	

	}
}