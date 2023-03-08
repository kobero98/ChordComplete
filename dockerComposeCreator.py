import sys

def create(Num_Nodi,Num_Repliche):
	numRep=str(Num_Repliche)
	file=open("Docker-compose.yaml","w")
	file.write("version: \"3.9\"\n")
	file.write("services:\n")
	file.write("  registration_server:\n")
	file.write("    hostname: register\n")
	file.write("    container_name: register\n")
	file.write("    build:\n")
	file.write("      context: .\n")
	file.write("      dockerfile: ./DockerFile/server_register\n")
	file.write("    ports:\n")
	file.write("      - 8000:8000\n")
	file.write("    environment:\n")
	file.write("      - REPLICHE="+numRep+"\n")
	file.write("    networks:\n")
	file.write("      - mynetwork\n")
	for i in range(1,Num_Nodi+1):
		for j in range(1,Num_Repliche+1):
			ID=str(8000+i*100+j)
			GID=str(8000+i)
			file.write("  nodo"+ID+":\n")
			file.write("    container_name: nodo"+ID+"\n")
			file.write("    hostname: nodo"+ID+"\n")
			file.write("    build:\n")
			file.write("      context: .\n")
			file.write("      dockerfile: ./DockerFile/nodo\n")
			file.write("    ports:\n")
			file.write("      - "+ID+":8005\n")
			file.write("    environment:\n")
			file.write("      - PORT_EXSPOST="+ID+"\n")
			file.write("      - CODICE_HASH="+GID+"\n")
			file.write("    depends_on:\n")
			file.write("      - registration_server\n")
			file.write("    networks:\n")
			file.write("      - mynetwork\n")
	file.write("networks:\n")
	file.write("  mynetwork:\n")
	file.write("   name: network1234\n")
z=int(sys.argv[1])
z2=int(sys.argv[2])
create(z,z2)