package simulation

import (
	"fmt"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestKubernetesProxy(t *testing.T) {
	cfg := DefaultKubernetesAdapterConfig()
	// Define k8s client configuration
	k8scfg, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		t.Fatalf("could not build config: %v", err)
	}

	server, err := NewProxyServer(k8scfg)

	l, err := server.Listen("127.0.0.1", 8888)
	if err != nil {
		t.Fatalf("failed to start proxy: %v", err)
	}
	defer l.Close()
	fmt.Printf("Listening on %s\n", l.Addr().String())
	server.ServeOnListener(l)

	/*err = NewProxyServer(k8scfg)
	if err != nil {
		t.Fatalf("meh: %v", err)
	}*/

}
