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
	contatoreRepliche int
}
var Num_Repl int
var lista_nodi []appoggio
var count = 0
var request = 0

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
	client, err := rpc.DialHTTP("tcp", nodeToContact.Ip[0]+":"+strconv.Itoa(nodeToContact.Port))
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
	for j<lista_nodi[i].contatoreRepliche{
		if lista_nodi[i].lista_Duplicati[j].PortExtern==nodo.PortExtern{
			return true
		}
		j=j+1
	}
	return false
}
//un nodo ti contatta se il leader non esiste gestisce lui la risorsa altrimenti diventa follower
func (t *Manager) Register(param *ParamRegister, reply *ReplyRegistration) error {
	fmt.Println("un nodo si é connesso", count, request)
	fmt.Println(*param)
	//aggiungere ulteriore controllo di veridicità del nodo
	i:=ControlloEsistenzaLeader(param.Nodo.Index) 
	if i!=-1 && lista_nodi[i].nodo.PortExtern!=param.Nodo.PortExtern {
		if controlloElementoGiaInserito(i,param.Nodo) {
			//se il nodo é già presente, invio un messaggio contenente le vecchie informazioni
			reply.Leader=lista_nodi[i].nodo
			reply.NumRep=lista_nodi[i].contatoreRepliche
			reply.IsReplica=true
			reply.Precedente, reply.Successivo = get_PrecSucc(param.Nodo)
			if lista_nodi[i].status==0{
				reply.IsReplica=false
				return nil
			}
		} else if lista_nodi[i].contatoreRepliche<Num_Repl{
			lista_nodi[i].lista_Duplicati[lista_nodi[i].contatoreRepliche]=param.Nodo
			fmt.Println("contatore repliche pre: ",lista_nodi[i].contatoreRepliche)
			lista_nodi[i].contatoreRepliche=lista_nodi[i].contatoreRepliche+1
			fmt.Println("contatore repliche post: ",lista_nodi[i].contatoreRepliche)
			//aggiungere comunicazione verso il leader
			reply.Leader=lista_nodi[i].nodo
			reply.NumRep=lista_nodi[i].contatoreRepliche
			reply.IsReplica=true
			reply.Precedente, reply.Successivo = get_PrecSucc(param.Nodo)
			if lista_nodi[i].status==0{
				reply.IsReplica=false
				return nil
			}
			if updateListaReplica(lista_nodi[i].nodo,param.Nodo)==-1 {
				lista_nodi[i].contatoreRepliche=lista_nodi[i].contatoreRepliche-1
			}
			return nil
		}else{
			//aggiungere cosa bisogna mettere in caso di saturazione delle repliche
			reply.IsReplica=true
			reply.Leader=param.Nodo
			return nil
		}
	}
	var a appoggio
	a.nodo = param.Nodo
	a.status = 0
	a.contatoreRepliche=0
	a.lista_Duplicati=make([]Node,Num_Repl)
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
	reply.IsReplica= false
	return nil
}
func (t *Manager) Unregister(node *Node, reply *Node) error {
	fmt.Println("un nodo si é disconnesso")
	count = count - 1
	reply = nil
	return nil
}

func heartBit() {
	for true {
		for i := 0; i < len(lista_nodi); i++ {
			if lista_nodi[i].status == 1 {
				client, err := rpc.DialHTTP("tcp", lista_nodi[i].nodo.Ip[0]+":"+strconv.Itoa(lista_nodi[i].nodo.Port))
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
		time.Sleep(10 * time.Second)
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
	go heartBit()
	http.Serve(l, nil)
	fmt.Println("fine programma in go")

}
