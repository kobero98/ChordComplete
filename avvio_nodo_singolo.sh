#! /bin/bash
C1=0
C2=0
y=1
z=8
while getopts x:y:z: flag 
do
    case "${flag}" in
        x) x=${OPTARG}
           C1=1
           ;;
        y) y=${OPTARG}
            C2=1
            ;;
        z) z=${OPTARG}
            ;;
    esac
done
if [[ "$C1" == "1" && "$C2" == "1" ]] 
then
    cd DockerFile
    docker build -f nodo -t kobnodo ../.
    cd ..
    echo "PORT_EXSPOST="$y>env.list
    echo "CODICE_HASH="$x>>env.list
    echo "BIT="$z>>env.list
    docker run -p $y:8005 -itd --name nodo$y --network network1234 --hostname nodo$y --env-file env.list kobnodo
else
    echo "bash avvio.sh -x [numero_nodo] [-y [IDENTIFICATIVO]] [-z [Dimensione dell'anello]]"
fi