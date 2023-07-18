# ChordComplete
Progetto finale per il corso di Sistemi Distribuiti e Cloud Computing della facoltà di Ingegneria Informatica Magistrale dell'università di Roma Tor vergata
# Requisiti
Il progetto richiede di avere sul proprio pc una versione di docker 3.9 o superiore.
# Obiettivo
Realizzare un sistema che implementi la rete di Chord, una rete strutturata, per memorizzare stringhe. In particolare, si sono implementati tutti i meccanismi di inserimento e eliminazione di un nodo, replicazione basata su primary-backup e algoritmo di elezione in caso di fallimento del nodo Leader.
# Come avviare il progetto
bisognerà eseguire il comando all'interno della cartella principale
```
bash avvio_Server.sh -x [Numero_di_nodi] [-y [Numero_di_repliche]] [-z [Numero_di_bit]]
```
inoltre, é disponibile uno script per poter aggiungere i nodi all'anello, e questi veranno inseriti anche lo start-up del sistema
```
bash avvio_nodo_singolo.sh -x [IDGroup] -y [IDNodo] [-z [Numero_di_bit]]
```
Verra instanziato un server_registry sulla porta 8000 e tanti nodi quanti specificati dal flag x sulle porte dalla [8000,...,8000+x]<br>
Inseguito si potrà accedere al nodo utilizzando il file eseguibile client in due modalità:
- put: si potrà memorizzare un valore nel sistema che restituirà la chiave dove é stato memorizzato
```
./client 1 string
```
- get: si potrà prendere il valore dal sistema passando come parametro la chiave
```
./client 0 key
```
- remove: si potrà elimanre la parola dal sistema passando come parametro la chiave
```
./client 2 key
```
