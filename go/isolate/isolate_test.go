package isolate

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	assert "github.com/stretchr/testify/require"
	"go.skia.org/infra/go/deepequal"
	"go.skia.org/infra/go/testutils"
)

func TestIsolateTasks(t *testing.T) {
	testutils.LargeTest(t)

	// Setup.
	workdir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer testutils.RemoveAll(t, workdir)

	c, err := NewClient(workdir, ISOLATE_SERVER_URL_FAKE)
	assert.NoError(t, err)

	ctx := context.Background()
	do := func(tasks []*Task, expectErr string) []string {
		hashes, err := c.IsolateTasks(ctx, tasks)
		if expectErr == "" {
			assert.NoError(t, err)
			assert.Equal(t, len(tasks), len(hashes))
			return hashes
		} else {
			assert.EqualError(t, err, expectErr)
		}
		return nil
	}

	// Write some files to isolate.
	writeIsolateFile := func(filepath string, contents *isolateFile) {
		f, err := os.Create(filepath)
		assert.NoError(t, err)
		defer testutils.AssertCloses(t, f)
		assert.NoError(t, contents.Encode(f))
	}

	myFile1 := "myfile1"
	myFile1Path := path.Join(workdir, myFile1)
	assert.NoError(t, ioutil.WriteFile(myFile1Path, []byte(myFile1), 0644))
	myFile2 := "myfile2"
	myFile2Path := path.Join(workdir, myFile2)
	assert.NoError(t, ioutil.WriteFile(myFile2Path, []byte(myFile2), 0644))

	dummyIsolate1 := path.Join(workdir, "dummy1.isolate")
	writeIsolateFile(dummyIsolate1, &isolateFile{
		Includes: []string{},
		Files:    []string{myFile1},
	})
	dummyIsolate2 := path.Join(workdir, "dummy2.isolate")
	writeIsolateFile(dummyIsolate2, &isolateFile{
		Includes: []string{},
		Files:    []string{myFile2},
	})

	// Empty tasks list.
	do([]*Task{}, "")

	// Invalid task.
	t1 := &Task{}
	_ = do([]*Task{t1}, "BaseDir is required.")
	t1.BaseDir = workdir
	_ = do([]*Task{t1}, "IsolateFile is required.")
	t1.IsolateFile = dummyIsolate1
	// Task with no OsType is ok.
	_ = do([]*Task{t1}, "")
	// Add OsType for below tests.
	t1.OsType = "linux"

	// Minimum valid task.
	hashes := do([]*Task{t1}, "")
	h1 := hashes[0]

	// Add a duplicate task.
	t2 := &Task{
		BaseDir:     workdir,
		IsolateFile: dummyIsolate1,
		OsType:      "linux",
	}
	hashes = do([]*Task{t1, t2}, "")
	deepequal.AssertDeepEqual(t, hashes, []string{h1, h1})

	// Tweak the second task.
	t2.IsolateFile = dummyIsolate2
	hashes = do([]*Task{t1, t2}, "")
	h2 := hashes[1]
	assert.NotEqual(t, h1, h2)
	deepequal.AssertDeepEqual(t, hashes, []string{h1, h2})

	// Add a dependency of t2 on t1. Ensure that we get a different hash,
	// which implies that the dependency was added successfully.
	t2.Deps = []string{h1}
	hashes = do([]*Task{t2}, "")
	assert.NotEqual(t, h2, hashes[0])

	// Isolate a bunch of tasks individually and then all at once, ensuring
	// that we get the same hashes in the correct order.
	tasks := []*Task{}
	expectHashes := []string{}
	for i := 0; i < 11; i++ {
		f := fmt.Sprintf("myfile%d", i)
		fp := path.Join(workdir, f)
		assert.NoError(t, ioutil.WriteFile(fp, []byte(f), 0644))
		dummyIsolate := path.Join(workdir, fmt.Sprintf("dummy%d.isolate", i))
		writeIsolateFile(dummyIsolate, &isolateFile{
			Includes: []string{},
			Files:    []string{f},
		})
		t := &Task{
			BaseDir:     workdir,
			IsolateFile: dummyIsolate,
			OsType:      "linux",
		}
		h := do([]*Task{t}, "")
		tasks = append(tasks, t)
		expectHashes = append(expectHashes, h[0])
	}
	gotHashes := do(tasks, "")
	deepequal.AssertDeepEqual(t, expectHashes, gotHashes)
}
