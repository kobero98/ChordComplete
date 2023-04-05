package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("esseguire il programma:")
		fmt.Println("./client 1 val per immagazzinare una stringa")
		fmt.Println("./client 0 key per ottenere la stringa su una determinata chiave")
		fmt.Println("./client 2 key per eliminare una chiave")
		return
	}
	x, errconv := strconv.Atoi(os.Args[1])
	if errconv != nil || (x != 0 && x != 1 && x!= 2) {
		log.Fatal("error valore non valido selezionare o 0 o 1 o 2 ")
		return
	}
	client, err := rpc.DialHTTP("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var result int
	result = 0
	var contact Node
	err = client.Call("Manager.ContactClient", &result, &contact)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	client.Close()
	fmt.Println("il nodo é: ", contact)
	var reply int
	var parola string
	client1, err1 := rpc.DialHTTP("tcp", /*contact.Name+*/"0.0.0.0:"+strconv.Itoa(contact.Port))
	if err1 != nil {
		log.Fatal("dialing:", err)
	}

	if x == 2 {
		reply, _ = strconv.Atoi(os.Args[2])
		var parola2 string
		err = client1.Call("ChordNode.Remove", &reply, &parola2)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Println("parola messa: ", parola2)
	}

	if x == 1 {
		parola = os.Args[2]
		fmt.Println(parola)
		err = client1.Call("ChordNode.Put", &parola, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Println("risposta: ", reply)
	}

	if x == 0 {
		reply, _ = strconv.Atoi(os.Args[2])
		var parola2 string
		err = client1.Call("ChordNode.Get", &reply, &parola2)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Println("parola messa: ", parola2)
	}
}
