package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/NiravShah1729/semdb/protocol"
	"github.com/NiravShah1729/semdb/store"
	"github.com/NiravShah1729/semdb/embedding"

)

func main() {
	kvStore := store.NewStore()
	emb := embedding.NewLocalEmbedder("http://localhost:8000")
	go kvStore.StartActiveEviction()

	listener,err := net.Listen("tcp",":8080")
	if err != nil{
		log.Fatalf("Failed to bind to port 8080: %v",err)
	}
	defer listener.Close()

	log.Println("Server listening on 8080")

	for {
		conn,err := listener.Accept()
		fmt.Println("Client connected")

		if err != nil{
			log.Printf("Error accepting connection: %v",err)
			continue
		}
		go handleConnection(conn,kvStore,emb)
	}
}

func handleConnection(conn net.Conn, kv *store.Store, emb embedding.Embedder){
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
		case "CLIENT":
    		writer.Write(protocol.Value{
        		Type: protocol.TypeSimpleString,
        		Str:  "OK",
    		})
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
			kv.Set(key,val,0)
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
		case "EXPIRE":
			if len(args) < 2 {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments",
				})
				continue
			}
			key := args[0].String()
			sec,err := strconv.Atoi(args[1].String())
			if err != nil {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong type of arguments",
				})
				continue
			}
			ok := kv.Expire(key,time.Duration(sec)*time.Second)
			if ok{
				writer.Write(protocol.Value{
					Type: protocol.TypeInteger,
					Num: 1,
				})
			}else{
				writer.Write(protocol.Value{
					Type: protocol.TypeInteger,
					Num: 0,
				})
			}
		case "TTL":
			if len(args) < 1{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments",
				})
				continue
			}
			key := args[0].String()
			ttl := kv.TTL(key)
			writer.Write(protocol.Value{
				Type: protocol.TypeInteger,
				Num : ttl,
			})

		//for hashes
		case "HSET":
			if len(args) < 3{
				writer.Write(protocol.Value{
					Type:protocol.TypeError,
					Str: "ERR wrong number of arguments in HSET command",
				})
				continue
			}
			key := args[0].String()
			field := args[1].String()
			val := args[2].String()

			res := kv.HSet(key,field,val)
			writer.Write(protocol.Value{
				Type: protocol.TypeInteger,
				Num: int64(res),
				
			})
		case "HGET":
			if len(args) < 2{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments for HGET command",
				})
				continue
			}
			key := args[0].String()
			field := args[1].String()

			val,ok := kv.HGet(key,field)
			if ok{
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(val),
				})
			}else{
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					IsNull: true,
				})
			}
		case "HDEL":
			if len(args) < 2 {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str:  "ERR wrong number of arguments for HDEL command",
				})
				continue
			}
			key := args[0].String()
			field := args[1].String()
			count := kv.HDel(key, field)
			writer.Write(protocol.Value{
				Type: protocol.TypeInteger,
				Num:  int64(count),
			})
		case "HGETALL":
			if len(args) < 1 {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str:  "ERR wrong number of arguments for HGETALL command",
				})
				continue
			}

			key := args[0].String()
			mp := kv.HGetAll(key)

			items := make([]protocol.Value, 0, len(mp)*2)
			for k, v := range mp {
				items = append(items, protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(k),
				})
				items = append(items, protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(v),
				})
			}
			writer.Write(protocol.Value{
				Type:  protocol.TypeArray,
				Array: items,
			})
		
		//for semantic 
		case "SETSEM":
			if len(args) < 3{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments",
				})
				continue
			}

			key := args[0].String()
			text := args[1].String()
			value := args[2].String()

			var ttl time.Duration
			if len(args) >= 4 {
				sec, err := strconv.Atoi(args[3].String())
				if err != nil {
					writer.Write(protocol.Value{
						Type: protocol.TypeError,
						Str:  "ERR invalid TTL value",
					})
					continue
				}
				ttl = time.Duration(sec) * time.Second
			}

			vector,err := emb.Embed(text)

			if err != nil{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR creating the vectore embedding",
				})
				continue
			}

			kv.Semantic.Add(key, text, value, vector, ttl)

			writer.Write(protocol.Value{
				Type: protocol.TypeSimpleString,
				Str: "OK",
			})
		case "GETSEM":
			if len(args) < 1{
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR wrong number of arguments",
				})
				continue
			}
			text := args[0].String()

			threshold := float32(0.80)
			if len(args) >= 2 {
				parsed, err := strconv.ParseFloat(args[1].String(), 32)
				if err != nil {
					writer.Write(protocol.Value{
						Type: protocol.TypeError,
						Str:  "ERR invalid threshold value",
					})
					continue
				}
				threshold = float32(parsed)
			}

			vector,err := emb.Embed(text)
			if err != nil {
				writer.Write(protocol.Value{
					Type: protocol.TypeError,
					Str: "ERR create the vectore for the string",
				})
				continue
			}

			val,ok := kv.Semantic.Search(vector, threshold)
			if !ok{
				writer.Write(protocol.Value{
					Type:protocol.TypeBulkString,
					IsNull: true,
				})
			}else{
				writer.Write(protocol.Value{
					Type: protocol.TypeBulkString,
					Bulk: []byte(val),
				})
			}

		case "CONFIG", "COMMAND", "SELECT":
			writer.Write(protocol.Value{
				Type: protocol.TypeSimpleString,
				Str:  "OK",
			})
		default:
			writer.Write(protocol.Value{
				Type: protocol.TypeError,
				Str:  fmt.Sprintf("ERR unknown command '%s'", cmd),
			})
		}
	

	}
}