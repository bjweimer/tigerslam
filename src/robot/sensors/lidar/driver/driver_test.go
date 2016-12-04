package driver

import (
    "testing"
    "time"
)

func TestConnection(t *testing.T) {
	u := MakeDefaultUrg()
	
	// Connect it
	defer u.Disconnect()
	err := u.Connect()
	if err != nil {
		t.Fatal(err)
	}
	
	// Check connection
	if got, want := u.IsConnected(), true; got != want {
		t.Errorf("Connected, but is not connected.")
	} else {
		t.Logf("Connected to LIDAR!")
	}
	
	// Get model name
	t.Logf("Model name: %s", u.GetModel())
	
	// Get data max
	t.Logf("DataMax: %d", u.dataMax)
}

func TestReceiveData(t *testing.T) {
	u := MakeDefaultUrg()
	
	n := 10
	
	// Connect
	defer u.Disconnect()
	err := u.Connect()
	if err != nil {
		t.Fatal(err)
	}
	
	// Request the data
	err = u.RequestInfiniteData()
	if err != nil {
		t.Fatal(err)
	}
	
	// Read data n times
	for i := 0; i < n; i++ {
		data, err := u.ReceiveData()
		if err != nil {
			t.Error(err)
		}
		t.Log(data)
	}
}