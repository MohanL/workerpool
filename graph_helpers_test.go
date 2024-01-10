package workerpool

import (
	"sync"
	"time"

	. "example.com/graphes"
)

func cloneGraph(node *Node, nodeNums int, goRoutines int) *Node {

	if node == nil {
		return nil
	}

	// fixed by queue size
	wp := NewWorkerPool(goRoutines, nodeNums*nodeNums, 10*time.Second)
	defer wp.Stop()

	head_clone := node.Clone()
	var visited sync.Map

	visited.Store(node.Val, head_clone)
	var wg sync.WaitGroup //use to synchronize all tasks

	task := func() (interface{}, error) {
		defer wg.Done()
		cloneNodeNeighboursChan, err := dfsCloneWithWorkerPool(node, head_clone, &visited, &wg, wp)
		return func() (interface{}, error) {
			defer wg.Done()
			for i := 0; i < len(head_clone.Neighbors); i++ {
				head_clone.Neighbors[i] = <-cloneNodeNeighboursChan
			}
			close(cloneNodeNeighboursChan)
			return nil, nil
		}, err
	}

	wg.Add(1)
	wp.Submit(task)
	wg.Wait()

loop:
	for {
		select {
		case i := <-wp.resultChan:
			if i.Error != nil {
				panic(i.Error)
			}
			if i.Result != nil {
				wg.Add(1)
				wp.Submit(i.Result.(func() (interface{}, error)))
			}
		default:
			// If no value ready, channel is drained and we can exit the loop
			// fmt.Println("Channel is drained")
			break loop
		}
	}

	wg.Wait()
	return head_clone
}

func dfsCloneWithWorkerPool(node *Node, cloneNode *Node, visited *sync.Map, wg *sync.WaitGroup, wp *WorkerPool) (chan *Node, error) {
	neighborsChan := make(chan *Node, len(node.Neighbors))
	for _, neighbor := range node.Neighbors {
		localNeighbor := neighbor
		wg.Add(1)
		wp.Submit(func() (interface{}, error) {
			defer wg.Done()
			neighborClone, loaded := visited.LoadOrStore(localNeighbor.Val, localNeighbor.Clone()) //cas
			neighborsChan <- neighborClone.(*Node)
			// // fmt.Printf("send neightbor clone %d to neighbour %d\n", neighborClone.(*Node).Val, node.Val)
			if !loaded {
				neighborCloneNeightborsChan, err := dfsCloneWithWorkerPool(localNeighbor, neighborClone.(*Node), visited, wg, wp)
				return func() (interface{}, error) {
					defer wg.Done()
					for i := 0; i < len(neighborClone.(*Node).Neighbors); i++ {
						neighborClone.(*Node).Neighbors[i] = <-neighborCloneNeightborsChan
					}
					close(neighborCloneNeightborsChan)
					return nil, nil
				}, err
			}
			return nil, nil
		})
	}
	return neighborsChan, nil
}
