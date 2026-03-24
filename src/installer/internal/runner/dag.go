package runner

import (
	"fmt"

	"dotfiles/installer/units"
)

// TaskNode wraps a Task with its parent context for global identification.
type TaskNode struct {
	GlobalID  string
	UnitID    string
	StageName string
	Task      units.Task
	Unit      *units.Unit
	Stage     *units.Stage
}

// DAG is a directed acyclic graph of TaskNodes across all units.
type DAG struct {
	nodes    map[string]*TaskNode
	edges    map[string][]string
	inDegree map[string]int
	waves    [][]*TaskNode
}

// NewDAG builds a global dependency graph from all units.
func NewDAG(u []units.Unit) (*DAG, error) {
	d := &DAG{
		nodes:    make(map[string]*TaskNode),
		edges:    make(map[string][]string),
		inDegree: make(map[string]int),
	}

	for i := range u {
		unit := &u[i]
		for j := range unit.Stages {
			stage := &unit.Stages[j]
			for _, task := range stage.Tasks {
				node := &TaskNode{
					GlobalID:  globalID(unit.ID, stage.Name, task.Name),
					UnitID:    unit.ID,
					StageName: stage.Name,
					Task:      task,
					Unit:      unit,
					Stage:     stage,
				}
				d.nodes[node.GlobalID] = node
				d.inDegree[node.GlobalID] = 0
			}
		}
	}

	for i := range u {
		unit := &u[i]
		for j := 1; j < len(unit.Stages); j++ {
			prev := &unit.Stages[j-1]
			curr := &unit.Stages[j]
			for _, prevTask := range prev.Tasks {
				fromID := globalID(unit.ID, prev.Name, prevTask.Name)
				for _, currTask := range curr.Tasks {
					toID := globalID(unit.ID, curr.Name, currTask.Name)
					d.addEdge(fromID, toID)
				}
			}
		}
	}

	for i := range u {
		unit := &u[i]
		for j := range unit.Stages {
			stage := &unit.Stages[j]
			for _, task := range stage.Tasks {
				taskID := globalID(unit.ID, stage.Name, task.Name)
				for _, dep := range task.DependsOn {
					depIDs := d.resolveDependency(dep, unit.ID)
					for _, depID := range depIDs {
						if _, ok := d.nodes[depID]; !ok {
							return nil, fmt.Errorf("task %q depends on %q which does not exist", taskID, depID)
						}
						d.addEdge(depID, taskID)
					}
				}
			}
		}
	}

	if err := d.computeWaves(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DAG) addEdge(from, to string) {
	d.edges[from] = append(d.edges[from], to)
	d.inDegree[to]++
}

func (d *DAG) resolveDependency(dep, currentUnitID string) []string {
	// Find the target unit's last stage from the Unit struct (ordered correctly).
	var lastStageName string
	for _, node := range d.nodes {
		if node.UnitID == dep && node.Unit != nil {
			stages := node.Unit.Stages
			if len(stages) > 0 {
				lastStageName = stages[len(stages)-1].Name
			}
			break
		}
	}
	if lastStageName == "" {
		return nil
	}

	// Return all tasks from the last stage of the dependency unit.
	var result []string
	for _, node := range d.nodes {
		if node.UnitID == dep && node.StageName == lastStageName {
			result = append(result, node.GlobalID)
		}
	}
	return result
}

func (d *DAG) computeWaves() error {
	inDeg := make(map[string]int, len(d.inDegree))
	for k, v := range d.inDegree {
		inDeg[k] = v
	}

	var queue []*TaskNode
	for id, deg := range inDeg {
		if deg == 0 {
			queue = append(queue, d.nodes[id])
		}
	}

	visited := 0
	for len(queue) > 0 {
		wave := make([]*TaskNode, len(queue))
		copy(wave, queue)
		d.waves = append(d.waves, wave)

		var next []*TaskNode
		for _, node := range queue {
			visited++
			for _, neighborID := range d.edges[node.GlobalID] {
				inDeg[neighborID]--
				if inDeg[neighborID] == 0 {
					next = append(next, d.nodes[neighborID])
				}
			}
		}
		queue = next
	}

	if visited != len(d.nodes) {
		return fmt.Errorf("cycle detected in dependency graph (%d/%d nodes visited)", visited, len(d.nodes))
	}

	return nil
}

// Waves returns the execution waves.
func (d *DAG) Waves() [][]*TaskNode {
	return d.waves
}

// Nodes returns all nodes in the DAG.
func (d *DAG) Nodes() map[string]*TaskNode {
	return d.nodes
}

func globalID(unitID, stageName, taskName string) string {
	return unitID + "/" + stageName + "/" + taskName
}
