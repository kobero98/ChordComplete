package main
import (
	"net/rpc"
	"log"
	"strconv"
)
func (t *ChordNode) ObtainNode(key *int, node *Node) error {
	//fmt.Println("mi hanno contattato per la chiave: ", *key)
	//fmt.Println("io mi occupo di: ", myPrecedente.Index, myNode.Index)
	if checkMyKey2(*key) {
		*node = myNode
		return nil
	} else {
		client, err := rpc.DialHTTP("tcp", mySuccessivo.Ip[0]+":"+strconv.Itoa(mySuccessivo.Port))
		if err != nil {
			log.Fatal("mynode ",myNode,"dialing:", err)
		}
		err = client.Call("ChordNode.ObtainNode", key, node)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	}
	return nil
}
//funzione che permette di impostare il precedente
func (t *ChordNode) Precedente(node *Node, reply *int) error {
	myPrecedente = *node
	*reply = 1
	return nil}
//funzione che permette di impostare il successivo
func (t *ChordNode) Successivo(node *Node, reply *int) error {
	mySuccessivo = *node
	return nil}
//funzione di HeartBit sostanzialmente non fa nulla 
func (t *ChordNode) HeartBit(answer *int, reply *int) error {
	*reply = 1
	return nil}
func (t *ChordNode) UpdateList(newNode *Node,reply *int) error{
	Lista_Eguali=append(Lista_Eguali,*newNode)
	return nil
} 