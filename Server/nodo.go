package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type ChordNode int
var myMap map[int]string
var FingerTable [8] Node //lista FingerTable
var Lista_Eguali [] Node  //lista Dei Nodi Replica
var mySuccessivo Node  //nodo successivo
var myPrecedente Node  //nodo precedente
//indirizzo del server register
var Addr_Server_register = "register"
var isReplica bool
var numRep int
var leader Node 
//variabile globale che rapresenta il nodo
var myNode Node
//funzione che contatta il server register e ritorna i nodi vicini
func init_registration() (Node, Node) {
	//effettuo la connessione al server register
	client, err := rpc.DialHTTP("tcp", Addr_Server_register+":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply ReplyRegistration
	var param ParamRegister
	param.Nodo=myNode
	err = client.Call("Manager.Register", &param, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Println(reply)
	isReplica=reply.IsReplica
	numRep=reply.NumRep
	leader=reply.Leader
	client.Close()
	return reply.Precedente, reply.Successivo
}
//funzione di inizializzazione dei nodi
func init_Node() {
	var err error
	myNode.Name, err = os.Hostname()
	addr, err := net.LookupHost(myNode.Name)
	if err != nil {
		log.Fatal("errore nel ottenere l'indirizzo ip dell'host:", err)
	}
	port := os.Getenv("PORT_EXSPOST")
	myNode.PortExtern, _ = strconv.Atoi(port)
	myNode.Ip = addr
	myNode.Port = 8005
	Indice_Hash:=os.Getenv("CODICE_HASH")
	myNode.Index = calcolo_hash(Indice_Hash)
	fmt.Println(myNode)
}
func PrintFingerTable(){
	fmt.Println("my node: ",myNode)
	fmt.Println("Stampo fingertable")
	for i:=0;i<len(FingerTable);i=i+1{
		fmt.Println(i,FingerTable[i])
	}
}
func init_fingerTable(){
	if mySuccessivo.Index==myNode.Index {
		return
	}
	var app Node
	var i int
	var key int
	i=0
	for app.Index != myNode.Index && i<8{
		key=(myNode.Index+1<<i) % 256
		client, err := rpc.DialHTTP("tcp", mySuccessivo.Ip[0]+":"+strconv.Itoa(mySuccessivo.Port))
		if err != nil {
			log.Fatal("mynode ",myNode,"dialing:", err)
		}
		err = client.Call("ChordNode.ObtainNode", &key, &app)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		if app.Index != myNode.Index{
			FingerTable[i].Name=app.Name
			FingerTable[i].Index=app.Index
			FingerTable[i].Port=app.Port
			FingerTable[i].PortExtern=app.PortExtern
			FingerTable[i].Ip=make([]string,1)
			FingerTable[i].Ip[0]=app.Ip[0]
		}
		i++;
	}
	return
}
//permette di tenere aggiornate e coerenti la fingertable ogni intervallo di tempo
func finger_Table(){
	fmt.Println("avvio finger_Table")
	for true {
		myPrecedente, mySuccessivo = init_registration()
		if myPrecedente.Name == "" && myPrecedente.Port == 0 {
		myPrecedente = myNode
		mySuccessivo = myNode
		init_fingerTable()
		} else {
			init_fingerTable()
			comunicationToPrecedente()
			comunicationToSuccessivo()
		}
		time.Sleep(10 * time.Second)
	}
	return
}
func comunicationToSuccessivo() {
	client, err := rpc.DialHTTP("tcp", mySuccessivo.Ip[0]+":"+strconv.Itoa(mySuccessivo.Port))
	if err != nil {
		log.Fatal("mynode ",myNode,"dialing:", err)
	}
	var reply int
	err = client.Call("ChordNode.Precedente", &myNode, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	client.Close()}
func comunicationToPrecedente() {
	client, err := rpc.DialHTTP("tcp", myPrecedente.Ip[0]+":"+strconv.Itoa(myPrecedente.Port))
	if err != nil {
		log.Fatal("mynode ",myNode,"dialing:", err)
	}
	var reply int
	err = client.Call("ChordNode.Successivo", &myNode, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	client.Close()}
func ChangeStatus() {
	client, err := rpc.DialHTTP("tcp", Addr_Server_register+":8000")
	if err != nil {
		log.Fatal("mynode ",myNode,"dialing:", err)
	}
	var reply int
	err = client.Call("Manager.ChangeStatus", &myNode, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	client.Close()
}
func heartBitReplica() bool {
	client, err := rpc.DialHTTP("tcp", leader.Ip[0]+":"+strconv.Itoa(leader.Port))
	if err != nil {
		return false
	}
	var reply int
	var answer int
	err = client.Call("ChordNode.HeartBit", &answer, &reply)
	if err != nil {
		client.Close()
		return false
	}
	client.Close()
	return true
}
func algoritmoBullyInverso(){
	return 
}
func main() {
	init_Node()
	myMap = make(map[int]string)
	myPrecedente, mySuccessivo = init_registration()
	if myPrecedente.Name == "" && myPrecedente.Port == 0 {
		fmt.Println("ciao 1")
		myPrecedente = myNode
		mySuccessivo = myNode
		init_fingerTable()

	} else {
		fmt.Println("ciao 2")
		init_fingerTable()
		comunicationToPrecedente()
		comunicationToSuccessivo()
	}
	fmt.Println("ciao")
	chord := new(ChordNode)
	rpc.Register(chord)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8005")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	ChangeStatus()
	go finger_Table()
	go http.Serve(l, nil)
	for true {
		if isReplica {
			fmt.Println("sono una replica")
			fmt.Println("con id: ",numRep)
			if !heartBitReplica() {
				fmt.Println("nodo leader crashuato")
				algoritmoBullyInverso()
			}
		}else{
			fmt.Println("sono il capo")
			fmt.Println("con id: ",numRep)
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("fine programma in go")
	return
}
