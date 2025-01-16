package main

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"os"
)

type Key string

type NodeAddress string

//const m = 6

//type Node struct {
//	Identifier  *big.Int
//	FingerTable []finger
//	Predecessor string
//	Successors  []string
//	IPAddress   string
//	FullAddress string //IP address and Channel Port
//	Bucket      map[Key]string
//}

//// CreateRing initializes the Chord Ring
//func (n *Node) CreateRing() {
//	n.Predecessor = nil
//	n.Successors = n
//	n.FingerTable = make([]*Node, m)
//}

//func (n *Node) Join(current *Node) {
//	n.Successors = current
//}

func (cr *ChordRing) ParseIP(args *InputArgs) {
	if args.IpAddr == "localhost" || args.IpAddr == "127.0.0.1" {
		args.IpAddr = "127.0.0.1"
		cr.IPAddress = args.IpAddr

	} else if args.IpAddr == "0.0.0.0" {
		getLocalIp := getLocalAddress()
		cr.IPAddress = getLocalIp
	} else if args.IpAddr == "public" {
		getLocalIp := GetPublicIP()
		cr.IPAddress = getLocalIp
	} else {
		cr.IPAddress = args.IpAddr
	}
}

func NewNode(args *InputArgs) *ChordRing {
	cr := &ChordRing{}
	cr.ParseIP(args)
	//Merge node's ip address and port into one variable
	IpPort := fmt.Sprintf("%s:%d", cr.IPAddress, args.Port)
	cr.FullAddress = IpPort
	//Initializing node's Identifier
	cr.Identifier = IdentifierGen(cr.FullAddress)

	//Initializing node's FingerTable
	cr.FingerTable = make([]finger, m+1)
	cr.FingerTableInit()

	//Initializing node Predecessor and a list of Successors
	cr.Predecessor = ""
	cr.SuccessorNum = args.SuccessorNum
	cr.Successors = make([]string, cr.SuccessorNum)
	cr.SuccessorInit()

	//Initialize the node's Bucket
	cr.Bucket = make(map[*big.Int]string)
	cr.KeyBackup = make(map[*big.Int]string)

	FolderName := FolderPathGen(cr.Identifier)
	// Create a directory for a node for saving the assigned files on it.
	err := os.MkdirAll(FolderName, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory")
	}
	cr.NodeFolder = FolderName

	//Generate a private key and a public key
	//cr.PublicKey, cr.PrivateKey, _ = GenAsymKeys("private.pem", "public.pem")
	cr.genRSAKey(2048)
	return cr
}

func (cr *ChordRing) SuccessorInit() {
	for i := 0; i < len(cr.Successors); i++ {
		cr.Successors[i] = cr.FullAddress
	}
}

// FingerTableInit initializes node n's FingerTable based on the formula of successor = (n.Identifier+ 2^(i-1)) mod 2^m
// Where i'th represents a finger in the table. In page 4 section IV.D 1 =< i =< m (m=6)
func (cr *ChordRing) FingerTableInit() {
	cr.FingerTable[0].Identifier = cr.Identifier
	cr.FingerTable[0].IPAddress = cr.IPAddress
	for i := 1; i < len(cr.FingerTable); i++ {
		addPart := jump(cr.IPAddress, i) //Hashing an IP address for given node
		cr.FingerTable[i].Identifier = addPart
		cr.FingerTable[i].IPAddress = cr.FullAddress
	}
}

func (cr *ChordRing) FindClosetFinger(id *big.Int) string {
	requestInfo := FindSucRequest{}
	requestInfo.InfoType = GetIP
	//fmt.Printf("FingerTable Size: %d\n", len(cr.FingerTable))
	for i := len(cr.FingerTable) - 1; i > 0; i-- {
		//sucReplay := CallFS(cr.FingerTable[i].IPAddress, "ChordRing.GetNodeInfo", &requestInfo)
		sucReplay := MakeCall[FindSucRequest, FindSucReplay](cr.FingerTable[i].IPAddress, "ChordRing.GetNodeInfo", requestInfo)
		SucId := IdentifierGen(sucReplay.SuccAddress)
		isBetween := between(cr.Identifier, SucId.Mod(SucId, hashMod), id, false)
		if isBetween {
			return cr.FingerTable[i].IPAddress
		}

	}
	return cr.Successors[0]
}

func (cr *ChordRing) GetNodeInfo(args *FindSucRequest, replay *FindSucReplay) error {
	switch args.InfoType {
	case GetIP:
		replay.SuccAddress = cr.FullAddress
	case GetID:
		replay.Identifier = cr.Identifier
	case GetSuc:
		replay.Successor = cr.Successors[0]
	case GetPre:
		replay.Predecessor = cr.Predecessor
	case GetSuccessors:
		replay.Successors = cr.Successors
	case GetPubKey:
		replay.PublicKey = cr.PublicKey
	case GetPriKey:
		replay.PrivateKey = cr.PrivateKey
	default:
		return errors.New("invalid info type")
	}
	return nil
}

func (cr *ChordRing) NodeServer() {
	err := rpc.Register(cr)
	if err != nil {
		fmt.Printf("Error with rpc Register %v:\n", err.Error())
	}
	//rpc.HandleHTTP()

	// sockname := coordinatorSock()
	// os.Remove(sockname)
	addr, err := net.ResolveTCPAddr("tcp", cr.FullAddress)

	if err != nil {
		log.Fatal("Inaccessible IP", err.Error())
	}
	listener, e := net.Listen("tcp", addr.String())
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Printf("NodeServer is running at %s\n", cr.FullAddress)
	go func(listener net.Listener) {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %s\n", err)
				break
			}
			go rpc.ServeConn(conn)

		}
	}(listener)
}

//func (cr *ChordRing) NodeServer() {
//	// Register the ChordRing object for RPC
//	err := rpc.Register(cr)
//	if err != nil {
//		fmt.Printf("Error with RPC Register: %v\n", err.Error())
//		return
//	}
//
//	// Load TLS certificates
//	cert, err := tls.LoadX509KeyPair("complete-cert.pem", "server-key.pem")
//	if err != nil {
//		log.Fatalf("Failed to load TLS certificate and key: %v", err)
//	}
//
//	// Create TLS configuration
//	tlsConfig := &tls.Config{
//		Certificates: []tls.Certificate{cert},
//		//MinVersion:   tls.VersionTLS12, // Enforce TLS 1.2 or higher
//		ClientAuth: tls.NoClientCert,
//	}
//
//	// Resolve TCP address
//	addr, err := net.ResolveTCPAddr("tcp", cr.FullAddress)
//	if err != nil {
//		log.Fatalf("Inaccessible IP: %v", err.Error())
//	}
//
//	// Create a TLS listener
//	listener, err := tls.Listen("tcp", addr.String(), tlsConfig)
//	if err != nil {
//		log.Fatalf("Failed to start TLS listener: %v", err)
//	}
//	defer listener.Close()
//
//	fmt.Printf("TLS NodeServer is running at %s\n", cr.FullAddress)
//
//	// Accept and serve connections
//	go func(listener net.Listener) {
//		for {
//			conn, err := listener.Accept()
//			if err != nil {
//				log.Printf("Error accepting connection: %s\n", err)
//				continue
//			}
//			go rpc.ServeConn(conn)
//		}
//	}(listener)
//}

func (cr *ChordRing) StoreFile(args *StoreFileArgs, replay *StoreFileReply) error {
	//cr.mutex.Lock()
	//defer cr.mutex.Unlock()
	key := args.Key
	switch args.StoreType {
	case MigrateUpload:
		uploadFolder := "../Upload_files"
		filePath := fmt.Sprintf("%s/%s", uploadFolder, args.FileName)
		if _, err := os.Stat(filePath); err == nil {
			openFile, err := os.Open(filePath)
			if err != nil {
				log.Fatal("At MigrateUpload: ", err)
			}
			nodeFolder := cr.NodeFolder
			fileContent, _ := io.ReadAll(openFile)
			file, err := os.Create(nodeFolder + "/" + args.FileName)
			if err != nil {
				log.Fatal("Error Creating a file (At MigrateUpload )", err)
			}
			_, err = file.Write(fileContent)
			if err != nil {
				log.Fatal("Error Writing a file (At MigrateUpload )", err)
			}
			//err := os.Rename(args.FileName, )
			//if err != nil {
			//	fmt.Printf("Error moving file to %s", nodeFolder+"/"+args.FileName)
			//	replay.IsSaved = false
			//}
			cr.Bucket[key] = args.FileName
			replay.IsSaved = true

		} else {
			//cr.Bucket[key] = args.FileName
			replay.IsSaved = false
		}
	case MigrateNode:
		nodeFolder := cr.NodeFolder
		for k, v := range args.MigratedBucket {

			//err := os.Rename(OldFileP, nodeFolder+"/"+v)
			//if err != nil {
			//	fmt.Printf("Error moving file to %s", nodeFolder+"/"+v)
			//}
			file, err := os.Create(nodeFolder + "/" + v)
			if err != nil {
				log.Fatal("Error Creating File: ", err)
				replay.IsSaved = false
			}
			decContent, err := rsa.DecryptPKCS1v15(rand.Reader, cr.PrivateKey, args.FileContent[k])
			if err != nil {
				log.Fatal("Error Decrypting File: ", err)
			}
			_, err = file.Write(decContent)

			if err != nil {
				log.Printf("Error writing to File: ", err)
			}
			//if args.Key.Cmp(k) == 0 {
			//	continue
			//} else {
			cr.Bucket[k] = v
			go func() {
				if err := RemoveFile(args.PrevNodeID, v); err != nil {
					log.Printf("Failed to remove file %s from node %s: %v", v, args.PrevNodeID, err)
				}
			}()
		}
		replay.IsSaved = true

	case KeyBackup:
		nodeFolder := cr.NodeFolder
		for k, v := range args.MigratedBucket {
			//if cr.IsExist(k) {
			//	continue
			//}
			file, err := os.Create(nodeFolder + "/" + v)
			if err != nil {
				log.Fatal("Error Creating Backup File (KeyBackup): ", err)
			}
			decContent, _ := rsa.DecryptPKCS1v15(rand.Reader, cr.PrivateKey, args.FileContent[k])
			//if err != nil {
			//	log.Printf("Error Decrypting Backup %s File's (KeyBackup): %v\n ", v, err)
			//	continue
			//}
			_, err = file.Write(decContent)
			if err != nil {
				log.Printf("Error writing to File (KeyBackup): ", err)
			}
			cr.KeyBackup[k] = v

			//go RemoveFile(args.PrevNodeID, v)

		}
	}
	//if file exists on the current directory /main, then move it to the node's folder

	return nil
}

// PutAll can be used to back up a bucket from a node that is about to leave.
func (cr *ChordRing) PutAll(args *BackupArgs, replay *BackupReply) error {
	expectedBucketSize := len(cr.Bucket) + len(args.Bucket)
	for k, v := range args.Bucket {
		cr.Bucket[k] = v
	}
	currentBucketSize := len(cr.Bucket)
	if currentBucketSize == expectedBucketSize {
		replay.IsSaved = true
	} else {
		replay.IsSaved = false
	}
	return nil
}

func (cr *ChordRing) MigrateBucket(moveAddress string) error {
	//
	getID := MakeCall[FindSucRequest, FindSucReplay](moveAddress, "ChordRing.GetNodeInfo",
		FindSucRequest{InfoType: GetID})
	newID := getID.Identifier
	MigratedBucket := make(map[*big.Int]string)
	ContentTable := make(map[*big.Int][]byte) //Declaring a map hash table, and its item is a list of byte
	for k, v := range cr.Bucket {
		fileKey := k
		fileName := v
		OldFilePath := FolderPathGen(cr.Identifier) + "/" + fileName
		openFile, err := os.Open(OldFilePath)
		if err != nil {
			log.Printf("Error opening file (MigrateBucket()) %s", OldFilePath)
		}
		defer openFile.Close()
		content, _ := io.ReadAll(openFile)
		encRep := MakeCall[FindSucRequest, FindSucReplay](moveAddress, "ChordRing.GetNodeInfo", FindSucRequest{InfoType: GetPubKey})

		//Encrypt the content of the file
		encContent, _ := rsa.EncryptPKCS1v15(rand.Reader, encRep.PublicKey, content)
		//getPredId := FindSucRequest{}
		//getPredId.InfoType = GetID
		//getIdRep := MakeCall[FindSucRequest, FindSucReplay](args.IPAddress, "ChordRing.GetNodeInfo", getPredId)

		//Check whether the successor node lies between k and the local node
		isBetween := between(cr.Identifier, fileKey, newID, true)
		if isBetween {

			MigratedBucket[k] = v
			ContentTable[k] = encContent
			delete(cr.Bucket, k)

		}
	}
	storeFileReq := StoreFileArgs{StoreType: MigrateNode, MigratedBucket: MigratedBucket,
		PrevNodeID: cr.Identifier, FileContent: ContentTable, isStored: false}

	storeRep := MakeCall[StoreFileArgs, StoreFileReply](moveAddress, "ChordRing.StoreFile", storeFileReq)

	if !storeRep.IsSaved {
		log.Printf("Error moving file to %v", newID)
	}
	return nil
}

func (cr *ChordRing) FindSuccessor(args *FindSucRequest, replay *FindSucReplay) error {
	//fmt.Printf("Joined Node ip: %s\n", args.IPAddress)
	//fmt.Printf("Joined Node id: %s\n", args.Identifier)
	//cr.mutex.Lock()
	//defer cr.mutex.Unlock()
	requestInfo := FindSucRequest{}
	requestInfo.InfoType = GetIP
	joinNodeID := args.Identifier.Mod(args.Identifier, hashMod)
	//newReplay := CallFS(cr.Successors[0], "ChordRing.GetNodeInfo", &requestInfo)
	newReplay := MakeCall[FindSucRequest, FindSucReplay](cr.Successors[0], "ChordRing.GetNodeInfo", requestInfo)

	idSuc := IdentifierGen(newReplay.SuccAddress)
	isBetween := between(cr.Identifier, joinNodeID, idSuc, true)
	if isBetween {
		//fmt.Printf("Found successor in between: %s\n", args.IPAddress)

		replay.SuccAddress = newReplay.SuccAddress

	} else { //Otherwise search on the finger table of this node
		sucAddress := cr.FindClosetFinger(idSuc)
		//sucReplay := CallFS(sucAddress, "ChordRing.FindSuccessor", args)
		newArgs := FindSucRequest{Identifier: joinNodeID}
		sucReplay := MakeCall[FindSucRequest, FindSucReplay](sucAddress, "ChordRing.FindSuccessor", newArgs)

		replay.SuccAddress = sucReplay.SuccAddress
	}
	//fmt.Printf("Found the successor: %s\n", replay.SuccAddress)
	//fmt.Printf("Found suc between %s\n", isBetween)
	return nil

}

// QuitChord handles the node's quit scenario, by backing up the bucket to the successor node
//func (cr *ChordRing) QuitChord() {
//	var SuccAddress string
//	if cr.Successors != nil && cr.Successors[0] != "" {
//		SuccAddress = cr.Successors[0]
//	} else {
//		findSucArgs := FindSucRequest{Identifier: cr.Identifier}
//		newReplay := MakeCall[FindSucRequest, FindSucReplay](cr.IPAddress, "ChordRing.FindSuccessor", findSucArgs)
//		SuccAddress = newReplay.SuccAddress
//	}
//	cr.MigrateBucket(SuccAddress)
//
//}

func (cr *ChordRing) Leave() error {
	// Notify successor about key migration
	MigratedBucket := make(map[*big.Int]string)
	FileContent := make(map[*big.Int][]byte)
	for key, value := range cr.Bucket {
		filePath := FolderPathGen(cr.Identifier) + "/" + value

		// Read file content
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Failed to open file %s: %v", filePath, err)
			continue
		}
		defer file.Close()

		content, err := io.ReadAll(file)

		getPubKey := MakeCall[FindSucRequest, FindSucReplay](cr.Successors[0], "ChordRing.GetNodeInfo", FindSucRequest{InfoType: GetPubKey})
		getPriKey := MakeCall[FindSucRequest, FindSucReplay](cr.Successors[0], "ChordRing.GetNodeInfo", FindSucRequest{InfoType: GetPriKey})
		fmt.Printf("Private Key %v\n", getPriKey.PrivateKey)
		fmt.Printf("Public Key %v\n", getPubKey.PublicKey)
		encContent, _ := rsa.EncryptPKCS1v15(rand.Reader, getPubKey.PublicKey, content)
		if err != nil {
			log.Printf("Failed to read file content for key %v: %v", key, err)
			continue
		}

		MigratedBucket[key] = value
		FileContent[key] = encContent
	}
	// Prepare request to store file on the successor
	req := StoreFileArgs{
		StoreType:      MigrateNode,
		MigratedBucket: MigratedBucket,
		FileContent:    FileContent,
		PrevNodeID:     cr.Identifier,
	}

	// Send file to the successor

	sucReplay := MakeCall[StoreFileArgs, StoreFileReply](cr.Successors[0], "ChordRing.StoreFile", req)
	if !sucReplay.IsSaved {
		log.Printf("Failed to migrate keys to successor %s!", cr.Successors[0])
	} else {
		// Remove key from local bucket
		for key, filePath := range MigratedBucket {
			delete(cr.Bucket, key)
			os.Remove(filePath)
		}

	}

	//// Notify the predecessor and successor to update their pointers
	//MakeCall[NotifyArgs, NotifyReply](cr.Successors[0], "ChordRing.NotifyRPC", NotifyArgs{
	//	NewAddress: cr.Predecessor,
	//})
	////if notifySuccessorErr != nil {
	////	log.Printf("Failed to notify successor: %v", notifySuccessorErr)
	////}
	//
	//MakeCall[NotifyArgs, NotifyReply](cr.Predecessor, "ChordRing.NotifyRPC", NotifyArgs{
	//	NewAddress: cr.Successors[0],
	//})
	//if notifyPredecessorErr != nil {
	//	log.Printf("Failed to notify predecessor: %v", notifyPredecessorErr)
	//}

	// Clear local state
	cr.Bucket = nil
	cr.Successors = nil
	return nil
}

// IsExist checks whether a given key is existed on the node's bucket or the backupkeys bucket
func (cr *ChordRing) IsExist(key *big.Int) bool {
	if _, ok := cr.Bucket[key]; ok {
		return true
	}
	if _, ok := cr.KeyBackup[key]; ok {
		return true
	}
	return false
}

func (cr *ChordRing) PrintState() {
	fmt.Println("++++++++++ Node Details ++++++++++")
	fmt.Printf("Node Address: %s\nNode Identifier: %v\nNode Predecessor: %s\n",
		cr.FullAddress,
		new(big.Int).SetBytes(cr.Identifier.Bytes()),
		cr.Predecessor)
	fmt.Println()
	fmt.Println("Node's Successors: ")
	for i, s := range cr.Successors {
		fmt.Printf("Successor %d: %s\n", i, s)
	}
	fmt.Println()
	fmt.Println("Finger Table: ")
	for i := 1; i < len(cr.FingerTable); i++ {
		fmt.Printf("Finger %d: %s, Address: %s\n", i, cr.FingerTable[i], cr.FingerTable[i].IPAddress)

	}
	fmt.Println()
	fmt.Println("Node bucket: ", cr.Bucket)

}
