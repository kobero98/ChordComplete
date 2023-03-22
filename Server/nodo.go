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
var FingerTable [] Node //lista FingerTable
var Lista_Eguali [] Node  //lista Dei Nodi Replica
var mySuccessivo Node  //nodo successivo
var myPrecedente Node  //nodo precedente
//indirizzo del server register
var Addr_Server_register = "register"
var isReplica bool
var numRep int
var Leader Node 
var NBit = 0
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
	isReplica=reply.IsReplica
	numRep=reply.NumRep
	Leader=reply.Leader
	client.Close()

	if isReplica == true{
		if Leader.PortExtern == myNode.PortExtern{
			log.Fatal("Errore Assegnazione nodo",isReplica)
			//return nil,nil
		}
	} else{
		if Leader.PortExtern != myNode.PortExtern{
			//in questo caso il nodo main non ha ancora cambiato il suo stato in attivo e dobbiamo aspettare che lo faccia
			//dormiamo per un tot e riavviamo la funzione
			time.Sleep(10 * time.Second)
			return init_registration()
			//return nil,nil
		}
	}
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
	NBit,_ = strconv.Atoi(os.Getenv("BIT"))
	FingerTable = make([]Node,NBit)
	myNode.Index = calcolo_hash(Indice_Hash)
	fmt.Println(myNode)
}
func PrintFingerTable(){
	fmt.Println("my node: ",myNode)
	fmt.Println("Stampo fingertable")
	fmt.Println(len(FingerTable))
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
	for app.Index != myNode.Index && i<NBit{
		key=(myNode.Index+1<<i) % 256
		client, err := rpc.DialHTTP("tcp", mySuccessivo.Ip[0]+":"+strconv.Itoa(mySuccessivo.Port))
		if err != nil {
			//log.Fatal("mynode ",myNode,"dialing:", err)
			fmt.Println("Errore connessione successivo",mySuccessivo,err)
			myPrecedente, mySuccessivo = init_registration()
			fmt.Println("ri ottentimento del precedente e del successivo")
			return
		}
		err = client.Call("ChordNode.ObtainNode", &key, &app)
		if err != nil {
			fmt.Println("Errore Obtain Node",mySuccessivo," node ID",key,err)
			client.Close()
			//log.Fatal("arith error:", err)
			return
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
		fmt.Println("errore creazione del client con il successivo ",err,myNode,mySuccessivo)
		//log.Fatal("mynode ",myNode,"dialing:", err)
		return
	}
	var reply int
	err = client.Call("ChordNode.Precedente", &myNode, &reply)
	if err != nil {
		fmt.Println("Errore comunicazione col successivo",err)
		client.Close()
		//log.Fatal("arith error:", err)
		return
	}
	client.Close()}
func comunicationToPrecedente() {
	client, err := rpc.DialHTTP("tcp", myPrecedente.Ip[0]+":"+strconv.Itoa(myPrecedente.Port))
	if err != nil {
		//log.Fatal("mynode ",myNode,"dialing:", err)
		fmt.Println("errore creazione del client con il precedente ",err,myNode,myPrecedente)
		return
	}
	var reply int
	err = client.Call("ChordNode.Successivo", &myNode, &reply)
	if err != nil {
		client.Close()
		fmt.Println("Errore comunicazione col successivo",err)
		//log.Fatal("arith error:", err)
		return
	}
	client.Close()
}
func ChangeStatus() {
	client, err := rpc.DialHTTP("tcp", Addr_Server_register+":8000")
	if err != nil {
		log.Fatal("mynode ",myNode,"dialing:", err)
	}
	var reply int
	err = client.Call("Manager.ChangeStatus", &myNode, &reply)
	if err != nil {
		client.Close()
		log.Fatal("arith error:", err)
	}
	client.Close()
}
func heartBitReplica() bool {
	client, err := rpc.DialHTTP("tcp", Leader.Ip[0]+":"+strconv.Itoa(Leader.Port))
	if err != nil {
		fmt.Println("il nodo é morto")
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
func algoritmoBullyInverso() bool {
	for i:=0;i<len(Lista_Eguali);i++{
		if Lista_Eguali[i].IDRep < numRep{
			client, err := rpc.DialHTTP("tcp", Lista_Eguali[i].Ip[0]+":"+strconv.Itoa(Lista_Eguali[i].Port))
			if err == nil {
				return false
			}
			var reply int
			var answer int
			err = client.Call("ChordNode.HeartBit", &answer, &reply)
			if err == nil {
				fmt.Println("l'algoritmo bully ha trovato qualcuno")
				client.Close()
				return false
			}
			client.Close()
		}
	}
	return true
}
func ChangeLeader() {
	client, err := rpc.DialHTTP("tcp", Addr_Server_register+":8000")
	if err != nil {
		log.Fatal("Contact register fail",err)
		return
	}
	var reply int
	err = client.Call("Manager.ChangeLeader", &myNode, &reply)
	if err != nil {
		log.Fatal("Fail call ChangeLeader",err)
		client.Close()
		return 
	}
	client.Close()
	for i:=0;i<len(Lista_Eguali);i++{
		if Lista_Eguali[i].Name!= "" {
			client, err = rpc.DialHTTP("tcp", Lista_Eguali[i].Ip[0]+":"+strconv.Itoa(Lista_Eguali[i].Port))
			if err != nil {
				log.Fatal("Contact node",Lista_Eguali[i],"index ",i,"fail",err)
				return
			}
			var reply2 int
			err = client.Call("ChordNode.NewLeader", &myNode, &reply2)
			if err != nil {
				log.Fatal("Fail call NewLeader",err)
				client.Close()
				return 
			}
			client.Close()
		}
	}
}
func takeMapData() {
	client, err := rpc.DialHTTP("tcp", Leader.Ip[0]+":"+strconv.Itoa(Leader.Port))
	if err != nil {
		log.Fatal("Contact Leader fail in takeMyData",err)
		return
	}

	err = client.Call("ChordNode.ObtainMapData", &myNode, &myMap)
	if err != nil {
		log.Fatal("Fail call ObtainMapData",err)
		client.Close()
		return 
	}
	client.Close()
}
func main() {
	init_Node()
	myMap = make(map[int]string)
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
	if isReplica {
		fmt.Println("Obtain Map")
		takeMapData()
		fmt.Println(myMap)
	}
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
			if heartBitReplica() == false {
				fmt.Println("nodo leader crashuato")
				if algoritmoBullyInverso() {
					ChangeLeader()
					isReplica=false
					fmt.Println("Ora dovrei diventare io il leader")
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("fine programma in go")
	return
}
