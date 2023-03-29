package main
import ("crypto/md5"
	"fmt"
	"log"
	"net/rpc"
	"strconv"
)
func checkKey(key int,precedente int,successivo int)bool{
	if successivo-precedente==0{
		return false
	}
	if successivo-precedente >0{
		return key >= precedente && key<successivo
	} else {
		return key >= precedente
	}

}

func nodeToContact(key int) Node{
	i:=0
	if checkKey(key,myNode.Index,FingerTable[0].Index){
		return FingerTable[0]
	}
	for i<NBit-1 {
		if checkKey(key,FingerTable[i].Index,FingerTable[i+1].Index){
			fmt.Println("indice ",i)
			return FingerTable[i]
		}
		i=i+1
	}
	return FingerTable[NBit-1]
}
func calcolo_hash(text string) int {
	hash := md5.Sum([]byte(text))
	x:=int(NBit/8)
	test:= make([]byte,x)
	var i int
	for i = 0; i < len(hash); i++ {
		test[i%x] = hash[i] ^ test[i%x]
	}
    result :=0
	for i =0; i < x; i++{
		result = result + int(test[i])<<(8*i)
	}
	return result
}
func checkMyKey2(key int) bool {
	//forse bastava fare se minore del myIndex.precede
	if myNode.Index-myPrecedente.Index == 0 {
		return true
	}
	if myNode.Index-myPrecedente.Index > 0 {
		return key > myPrecedente.Index && key <= myNode.Index
	}
	return key > myPrecedente.Index || key <= myNode.Index
}

//funzione che rimuove una chiave dal nodo
//ci devo pensare un attimo
func (t *ChordNode) Remove(key *int, reply *string) error {
	fmt.Println("mi hanno contattato per rimuove la kiave: ", *key)
	fmt.Println("io mi occupo di: ", myPrecedente.Index, myNode.Index)
	if checkMyKey2(*key) {
		str := myMap[*key]
		*reply = str
		return nil
	} else {
		client, err := rpc.DialHTTP("tcp", mySuccessivo.Name+":"+strconv.Itoa(mySuccessivo.Port))
		if err != nil {
			log.Fatal("dialing:", err)
		}
		err = client.Call("ChordNode.Get", key, reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	}
	return nil
}
//funzione che prende una risorsa in base alla chiave
func (t *ChordNode) Get(key *int, reply *string) error {
	if isReplica {
		return nil
	}
	fmt.Println("mi hanno contattato per la chiave: ", *key)
	fmt.Println("io mi occupo di: ", myPrecedente.Index, myNode.Index)
	if checkMyKey2(*key) {
		str, test := myMap[*key]
		if test == false {
			str = "NOVALUE"
			return nil
		}
		*reply = str
		return nil
	} else {
		appNode:=nodeToContact(*key)
		fmt.Println(appNode)
		client, err := rpc.DialHTTP("tcp", appNode.Name+":"+strconv.Itoa(appNode.Port))
		if err != nil {
			myPrecedente, mySuccessivo = init_registration()
			fmt.Println("ri ottentimento del precedente e del successivo")
		}
		err = client.Call("ChordNode.Get", key, reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	}
	return nil
}
func (t *ChordNode) UpdateReplica(param *ParamUpdateReplica,reply *int) error {
	myMap[param.Key]=param.Parola
	return nil
}
func updateReplicaBase(key int, parola string){
	i:=0
	for i<len(Lista_Eguali){
		client, err := rpc.DialHTTP("tcp", Lista_Eguali[i].Name+":"+strconv.Itoa(Lista_Eguali[i].Port))
		if err != nil {
			//gestire errore uno dei nodi crasha
			fmt.Println("Uno dele repliche non é più up ",err,)
			continue
		}
		var param ParamUpdateReplica
		param.Key=key
		param.Parola=parola
		var reply int
		err = client.Call("ChordNode.UpdateReplica", &param, &reply)
		if err != nil {
			fmt.Println("Errore Update Replica",Lista_Eguali[i],err)
			return 
		}
		client.Close()
		i=i+1
	}
}
//funzione che mette nell'anello una stringa
func (t *ChordNode) Put(parola *string, reply *int) error {
	if isReplica {
		*reply=-1
		return nil
	}
	fmt.Println("mi hanno contattato per la parola: ", *parola)
	key := calcolo_hash(*parola)
	fmt.Println("la chiave  é ", key)
	fmt.Println("io mi occupo di: ", myPrecedente.Index+1, myNode.Index)
	if checkMyKey2(key) {
		updateReplicaBase(key,*parola)
		myMap[key] = *parola
	} else {
		//PrintFingerTable()
		appNode:=nodeToContact(key)
		fmt.Println("Nodo scelto: ",appNode," IP:",appNode.Ip[0],"port: ",appNode.Port)
		client, err := rpc.DialHTTP("tcp", appNode.Name+":"+strconv.Itoa(appNode.Port))
		if err != nil {
			myPrecedente, mySuccessivo = init_registration()
			fmt.Println("ri ottentimento del precedente e del successivo")
			return err
		}
		err = client.Call("ChordNode.Put", parola, reply)
		if err != nil {
			myPrecedente, mySuccessivo = init_registration()
			log.Fatal("arith error:", err)
			fmt.Println("ri ottentimento del precedente e del successivo")
			return err
		}
	}
	*reply = key
	return nil
}

