package zfile

import (
	"bufio"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestMain(m *testing.M) {
	m.Run()
	_ = Remove("6.txt")
	_ = Remove("7.txt")
	_ = Remove("8.txt")
}

func TestMemoryFile(t *testing.T) {
	tt := zlsgo.NewTest(t)
	f := NewMemoryFile("6.txt", MemoryFileAutoFlush(1), MemoryFileFlushBefore(func(f *MemoryFile) error {
		f.SetName("7.txt")
		t.Log(f.Size())
		return nil
	}))

	t.Log(f.Name())
	tt.EqualTrue(!f.IsDir())
	t.Log(f.ModTime())
	t.Log(f.Mode())
	t.Log(f.Sys())
	t.Log(f.Size())

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			_, err := f.Write([]byte(strconv.Itoa(i) + "\n"))
			wg.Done()
			tt.ErrorNil(err)
		}(i)
	}

	b := []byte("--\n")
	_, err := f.Write(b)
	tt.ErrorNil(err)
	t.Log(len(f.Bytes()))
	wg.Wait()
	t.Log(len(f.Bytes()))
	tt.ErrorNil(f.Close())
	t.Log(FileSize("7.txt"))
	tt.ErrorNil(f.Close())
}

func BenchmarkFileMem6(b *testing.B) {
	name := "6.txt"
	f := NewMemoryFile(name)
	for i := 0; i < b.N; i++ {
		_, err := f.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Fatal(err)
		}
	}
	WriteFile(name, f.Bytes())
}

func BenchmarkFileReal8(b *testing.B) {
	name := "8.txt"
	for i := 0; i < b.N; i++ {
		err := WriteFile(name, []byte(strconv.Itoa(i)), true)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileBufio7(b *testing.B) {
	name := "7.txt"
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0766)
	write := bufio.NewWriter(file)
	for i := 0; i < b.N; i++ {
		_, err := write.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Fatal(err)
		}
	}
	write.Flush()
}
