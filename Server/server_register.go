package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
	"os"
)
type Manager int
type appoggio struct {
	nodo   Node
	status int
	lista_Duplicati []Node
	lista_Indici_Liberi []int
	contatoreRepliche int
}
var Num_Repl int
var lista_nodi []appoggio
var count = 0
var request = 0
var NBit = 0

func printList() {
	fmt.Println("dimensione lista", len(lista_nodi))
	for i := 0; i < len(lista_nodi); i++ {
		fmt.Println(lista_nodi[i].nodo)
	}
}
func remove_elemento(n appoggio) int {
	if len(lista_nodi) == 0 {
		return 0
	}
	for i := 0; i < len(lista_nodi); i++ {
		if lista_nodi[i].nodo.Index == n.nodo.Index {
			lista_nodi = append(lista_nodi[:i], lista_nodi[i+1:]...)
			return 0
		}
	}
	return -1
}
func add_elemento(n appoggio) int {
	if len(lista_nodi) == 0 {
		lista_nodi = append(lista_nodi, n)
		return 0
	}
	var app []appoggio
	app = append(app, n)
	index := -1
	for i := 0; i < len(lista_nodi); i++ {
		if lista_nodi[i].nodo.Index == n.nodo.Index {
			if lista_nodi[i].nodo.Name == n.nodo.Name && lista_nodi[i].nodo.Port == n.nodo.Port && lista_nodi[i].nodo.Ip[0] == n.nodo.Ip[0] {
				return -1
			}
			return -2
		}
		if lista_nodi[i].nodo.Index > n.nodo.Index {
			index = i
			break
		}
	}
	if index == 0 {
		lista_nodi = append(app, lista_nodi...)
		fmt.Println("dimensione lista", len(lista_nodi))
		return 0
	}
	if index == -1 {
		lista_nodi = append(lista_nodi, app...)
		fmt.Println("dimensione lista", len(lista_nodi))
		return 0
	}
	app = append(app, lista_nodi[index:]...)
	lista_nodi = append(lista_nodi[:index], app...)
	fmt.Println("dimensione lista", len(lista_nodi))
	return 0
}
func get_PrecSucc(n Node) (Node, Node) {
	var prec Node
	var succ Node
	for i:=0;i<len(lista_nodi);i++{
		if n.Index == lista_nodi[i].nodo.Index{
			j:=i-1
			if j<0 { j=j+len(lista_nodi) }
			for j != i {
				if lista_nodi[j].status == 1 {
					prec=lista_nodi[j].nodo
					break;
				}
				j=j-1
				if j<0 {j=j+len(lista_nodi)}
			}
			j=(i+1)%len(lista_nodi)
			for j!=i{
				if lista_nodi[j].status == 1 {
					succ=lista_nodi[j].nodo
					break;
				}
				j=(j+1)%len(lista_nodi)
			}
		}
	}
	return prec,succ
	/*
	for i := 0; i < len(lista_nodi); i++ {
		if n.Index == lista_nodi[i].nodo.Index {
			if i == 0 {
				return lista_nodi[len(lista_nodi)-1].nodo, lista_nodi[(i+1)%len(lista_nodi)].nodo
			}
			//fmt.Println("valore lista ", (i-1)%len(lista_nodi), (i+1)%len(lista_nodi))
			return lista_nodi[(i-1)%len(lista_nodi)].nodo, lista_nodi[(i+1)%len(lista_nodi)].nodo
		}
	}
	return n, n
	*/
}
func (t *Manager) ChangeStatus(node *Node, reply *int) error {
	//fmt.Println("ricevuto un cambio di stato dal nodo: ", node.Index)
	for i := 0; i < len(lista_nodi); i++ {
		if lista_nodi[i].nodo.Index == node.Index {
			lista_nodi[i].status = 1
			*reply = 0
			return nil
		}
	}
	*reply = -1
	return nil
}
func ControlloEsistenzaLeader(index int) int {
	i:=0
	for i<len(lista_nodi){
		if lista_nodi[i].nodo.Index == index{
			return i
		}
		i=i+1
	}
	return -1
}
func updateListaReplica(nodeToContact Node,nodeToPass Node) int {
	client, err := rpc.DialHTTP("tcp", nodeToContact.Name+":"+strconv.Itoa(nodeToContact.Port))
	if err != nil {
		fmt.Println("ri-ottentimento del precedente e del successivo")
		return -1
	}
	var x int
	err = client.Call("ChordNode.UpdateList", &nodeToPass, &x)
	if err != nil {
		//log.Fatal("arith error:", err)
		fmt.Println("ri ottentimento del precedente e del successivo")
		return -1
	}
	return 0
}
func controlloElementoGiaInserito(i int,nodo Node) bool {
	j:=0
	for j<len(lista_nodi[i].lista_Duplicati){
		if lista_nodi[i].lista_Duplicati[j].PortExtern==nodo.PortExtern{
			return true
		}
		j=j+1
	}
	return false
}
func controlloIndiceLibero(i int) int {
	for j:=0;j<len(lista_nodi[i].lista_Indici_Liberi);j++{
		if lista_nodi[i].lista_Indici_Liberi[j]!=-1{
			return j
		}
	}
	return -1
}
//un nodo ti contatta se il leader non esiste gestisce lui la risorsa altrimenti diventa follower
func (t *Manager) Register(param *ParamRegister, reply *ReplyRegistration) error {
	//aggiungere ulteriore controllo di veridicità del nodo
	i:=ControlloEsistenzaLeader(param.Nodo.Index) 
	if i!=-1 && lista_nodi[i].nodo.PortExtern!=param.Nodo.PortExtern {
		//se é vero questo allora abbiamo già un main node e questa é chiaramente una replica
		if controlloElementoGiaInserito(i,param.Nodo) {
			//se il nodo é già presente, invio un messaggio contenente le vecchie informazioni
			reply.Leader=lista_nodi[i].nodo
			k:=-1
			for j:=0;j<len(lista_nodi[i].lista_Duplicati);j++{
				if lista_nodi[i].lista_Duplicati[j].PortExtern == param.Nodo.PortExtern{
					k=j
				}
			}
			reply.NumRep=k
			reply.IsReplica=true
			reply.Precedente, reply.Successivo = get_PrecSucc(param.Nodo)
			if lista_nodi[i].status==0 {
				reply.IsReplica=false
				return nil
			}
		} else {
			lib:=controlloIndiceLibero(i)
			if lib!=-1{
				param.Nodo.IDRep=lib
				lista_nodi[i].lista_Duplicati[lib]=param.Nodo
				lista_nodi[i].lista_Indici_Liberi[lib]=-1
				//aggiungere comunicazione verso il leader
				reply.Leader=lista_nodi[i].nodo
				reply.NumRep=lib
				reply.IsReplica=true
				reply.Precedente, reply.Successivo = get_PrecSucc(param.Nodo)
				if lista_nodi[i].status==0{
					reply.IsReplica=false
					return nil
				}
				if updateListaReplica(lista_nodi[i].nodo,param.Nodo)==-1 {
					lista_nodi[i].lista_Indici_Liberi[lib]=lib
				}
				return nil
			}else{
				//aggiungere cosa bisogna mettere in caso di saturazione delle repliche
				reply.IsReplica=true
				reply.Leader=param.Nodo
				return nil
			}
		}
	}

	var a appoggio
	a.nodo = param.Nodo
	a.nodo.IDRep = 0
	a.status = 0
	a.contatoreRepliche=0
	a.lista_Duplicati=make([]Node,Num_Repl)
	a.lista_Indici_Liberi=make([]int,Num_Repl)
	for indice_libero:=0;indice_libero<Num_Repl;indice_libero++{
		a.lista_Indici_Liberi[indice_libero]=indice_libero
		}
	contr := add_elemento(a)
	if contr == -2 {
		reply = nil
		return nil
	}
	count = count + 1 + contr
	if count == 1 {
		reply = nil
		return nil
	}
	reply.Precedente, reply.Successivo = get_PrecSucc(param.Nodo)
	reply.NumRep = 0
	reply.Leader = param.Nodo
	reply.IsReplica= false
	return nil
}
// non utilizzata per ora
func (t *Manager) Unregister(node *Node, reply *Node) error {
	fmt.Println("un nodo si é disconnesso")
	count = count - 1
	reply = nil
	return nil
}
//funzione al momento da rivedere
func heartBit() {
	for true {
		for i := 0; i < len(lista_nodi); i++ {
			if lista_nodi[i].status == 1 {
				client, err := rpc.DialHTTP("tcp", lista_nodi[i].nodo.Name+":"+strconv.Itoa(lista_nodi[i].nodo.Port))
				if err != nil {
					fmt.Println("Elemento rimosso 1")
					remove_elemento(lista_nodi[i])
					continue
				}
				var reply int
				var answer int
				err = client.Call("ChordNode.HeartBit", &answer, &reply)
				if err != nil {
					fmt.Println("Elemento rimosso 2")
					remove_elemento(lista_nodi[i])
					client.Close()
					continue
				}
				client.Close()
			}
		}
		time.Sleep(2 * time.Second)
	}
}

var valore = 0

func (t *Manager) ContactClient(value *int, reply *Node) error {
	fmt.Println("mi hanno contattato")
	if len(lista_nodi) < 1 {
		reply = nil
		return nil
	} else {
		var index = valore % len(lista_nodi)
		*&reply.Name = lista_nodi[index].nodo.Name
		*&reply.Port = lista_nodi[index].nodo.PortExtern
		*&reply.Ip = lista_nodi[index].nodo.Ip
		*&reply.PortExtern = lista_nodi[index].nodo.PortExtern
		return nil
	}
}
func (t *Manager) ChangeLeader(newNode *Node,reply *int)error{
	fmt.Println("cambio Leader")
	i:=ControlloEsistenzaLeader(newNode.Index)
	if i== -1 {
		fmt.Println("ci devo pensare")
	}else{
		var newLeader Node
		newLeader.Name=newNode.Name
		newLeader.Port=newNode.Port
		newLeader.PortExtern=newNode.PortExtern
		newLeader.Index=newNode.Index
		newLeader.Ip=newNode.Ip
		newLeader.IDRep=newNode.IDRep
		lista_nodi[i].nodo=newLeader
		for j:=0;j<len(lista_nodi[i].lista_Duplicati);j++{
			if lista_nodi[i].lista_Duplicati[j].PortExtern == newNode.PortExtern {
				lista_nodi[i].lista_Indici_Liberi[j]=j
				lista_nodi[i].lista_Duplicati[j].Name=""
				lista_nodi[i].lista_Duplicati[j].Index=0
				lista_nodi[i].lista_Duplicati[j].Port=0
				lista_nodi[i].lista_Duplicati[j].PortExtern=0
			}
		}
		lista_nodi[i].status=1
	}
	return nil
}
func threadStampa() {
	for true {
		//fmt.Println(lista_nodi)
		printList()
		time.Sleep(5 * time.Second)
	}
}
func main() {
	Num_Repl,_ =strconv.Atoi(os.Getenv("REPLICHE")) 
	fmt.Println("inizio programma in go")
	lista_nodi = make([]appoggio, 0)
	manage := new(Manager)
	rpc.Register(manage)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8000")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go threadStampa()
	go heartBit()
	http.Serve(l, nil)
	fmt.Println("fine programma in go")

}
