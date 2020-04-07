package pss

func TestAPITopic(t *testing.T) { 
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	_, err = rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}

	// using the same key in for underlay address as well as it is not important for test
	baseAddress := network.NewBzzAddr(baseKey, baseKey)
	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	netStore := storage.NewNetStore(localStore, baseAddress)
	
	
	//Mock the store
	//call send
	//verify store, that trojan chunk has been stored correctly
}

