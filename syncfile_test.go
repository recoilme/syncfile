package syncfile

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

var sf *SyncFile

func ch(err error) {
	if err != nil {
		panic(err)
	}
}
func TestAppend(t *testing.T) {
	var err error
	file := "la.tmp"
	os.Remove(file)
	sf, err = NewSyncFile(file, 0666)
	defer sf.Close()
	ch(err)

	sf.Append([]byte("1234567890"))
	messages := make(chan int)
	readmessages := make(chan string)
	var wg sync.WaitGroup

	append := func(i int) {
		defer wg.Done()
		s := " " + strconv.Itoa(i) + " "
		sf.Append([]byte(s))
		messages <- i
	}

	read := func(i int) {
		defer wg.Done()

		b, err := sf.Read(10, 0)
		//content, err := ioutil.ReadFile(file)
		if err != nil {
			t.Error(err)
		}
		if string(b) != "1234567890" {
			t.Error("not seek to beginning of file with Rlock")
		}
		readmessages <- fmt.Sprintf("read N:%d  content:%s", i, string(b))
	}

	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go append(i)
		wg.Add(1)
		go read(i)
	}

	go func() {
		for i := range messages {
			_ = i
			//fmt.Println(i)
		}
	}()

	go func() {
		for i := range readmessages {
			fmt.Println(i)
		}
	}()

	wg.Wait()

}

func TestReadFile(t *testing.T) {
	var err error
	file := "tst.tmp"
	os.Remove(file)
	sf, err = NewSyncFile(file, 0666)
	if err != nil {
		t.Error(err)
	}
	sf.Append([]byte("1234567890"))
	sf.Append([]byte("\n1234567890"))
	var data []byte
	data, err = sf.ReadFile()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("data:", string(data))
	defer sf.Close()
}

func TestReadAppend(t *testing.T) {
	var err error
	file := "ra.tmp"
	os.Remove(file)
	sf, err = NewSyncFile(file, 0666)
	defer sf.Close()
	ch(err)

	sf.Append([]byte("abc"))
	messages := make(chan int)
	readmessages := make(chan string)
	var wg sync.WaitGroup

	append := func(i int) {
		defer wg.Done()
		s := " " + strconv.Itoa(i) + " "
		sf.Append([]byte(s))
		messages <- i
	}

	read := func(i int) {
		defer wg.Done()

		b, err := sf.ReadFile()
		//content, err := ioutil.ReadFile(file)
		if err != nil {
			t.Error(err)
		}

		readmessages <- fmt.Sprintf("read N:%d  content:%s", i, string(b))
	}

	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go append(i)
		wg.Add(1)
		if i == 10 {
			time.Sleep(2 * time.Second)
		}
		go read(i)
	}

	go func() {
		for i := range messages {
			//_ = i
			fmt.Println(i)
		}
	}()

	go func() {
		for i := range readmessages {
			fmt.Println(i)
		}
	}()

	wg.Wait()

}

/*
//no errors (
func TestWriteAt(t *testing.T) {

	file := "wa.tmp"
	//os.Remove(file)
	f, _ := NewSyncFile(file, 0666)
	defer f.Close()
	//for i := 0; i < 119; i++ {
	i := 200
	j := []byte(strconv.Itoa(i))
	//if i < 10 {
	fmt.Println(f.WriteAt(j, int64(i)))
	//	} else {
	//	fmt.Println(f.WriteAt(j, int64(i/10)))
	//}
	//}
}
*/
