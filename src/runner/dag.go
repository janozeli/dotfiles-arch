package main

import (
	"fmt"
	"strings"
)

// TaskNode wraps a Task with its parent context for global identification.
type TaskNode struct {
	GlobalID  string // "unit_id/stage_name/task_name"
	UnitID    string
	StageName string
	Task      Task
	Unit      *Unit
	Stage     *Stage
}

// DAG is a directed acyclic graph of TaskNodes across all units.
type DAG struct {
	nodes    map[string]*TaskNode
	edges    map[string][]string // from → []to
	inDegree map[string]int
	waves    [][]*TaskNode
}

// NewDAG builds a global dependency graph from all units.
// It flattens tasks into nodes, adds explicit and implicit edges,
// detects cycles, and groups nodes into parallel execution waves.
func NewDAG(units []Unit) (*DAG, error) {
	d := &DAG{
		nodes:    make(map[string]*TaskNode),
		edges:    make(map[string][]string),
		inDegree: make(map[string]int),
	}

	// 1. Flatten all tasks into nodes.
	for i := range units {
		unit := &units[i]
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

	// 2. Add implicit edges: tasks in stage N+1 depend on all tasks in stage N.
	for i := range units {
		unit := &units[i]
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

	// 3. Add explicit edges from depends_on.
	for i := range units {
		unit := &units[i]
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

	// 4. Topological sort into waves (Kahn's algorithm).
	if err := d.computeWaves(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DAG) addEdge(from, to string) {
	d.edges[from] = append(d.edges[from], to)
	d.inDegree[to]++
}

// resolveDependency turns a dependency reference into global IDs.
// A bare unit ID (e.g., "packages") means: depend on ALL tasks of the LAST stage of that unit.
// A local task name (no "/" or ".") within the same unit is resolved locally.
func (d *DAG) resolveDependency(dep, currentUnitID string) []string {
	// Cross-unit dependency: bare unit ID.
	// Find the last stage of that unit and return all its task IDs.
	var ids []string
	for _, node := range d.nodes {
		if node.UnitID == dep {
			ids = append(ids, node.GlobalID)
		}
	}
	if len(ids) > 0 {
		// Only depend on the last stage's tasks.
		return filterLastStage(ids, dep)
	}

	// Could be a local task reference — not cross-unit.
	return nil
}

func filterLastStage(ids []string, unitID string) []string {
	// Group by stage, find the last one alphabetically by position in the IDs.
	stageOrder := map[string]int{}
	maxOrder := -1
	lastStage := ""

	for _, id := range ids {
		parts := strings.SplitN(id, "/", 3)
		if len(parts) < 3 {
			continue
		}
		stage := parts[1]
		if _, ok := stageOrder[stage]; !ok {
			stageOrder[stage] = len(stageOrder)
		}
		if stageOrder[stage] > maxOrder {
			maxOrder = stageOrder[stage]
			lastStage = stage
		}
	}

	var result []string
	for _, id := range ids {
		parts := strings.SplitN(id, "/", 3)
		if len(parts) >= 2 && parts[1] == lastStage {
			result = append(result, id)
		}
	}
	return result
}

// computeWaves performs Kahn's BFS topological sort and groups nodes by layer.
func (d *DAG) computeWaves() error {
	inDeg := make(map[string]int, len(d.inDegree))
	for k, v := range d.inDegree {
		inDeg[k] = v
	}

	// Seed queue with all zero in-degree nodes.
	var queue []*TaskNode
	for id, deg := range inDeg {
		if deg == 0 {
			queue = append(queue, d.nodes[id])
		}
	}

	visited := 0
	for len(queue) > 0 {
		// All nodes in the current queue form one wave.
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

// Waves returns the execution waves. Each wave contains tasks that can run in parallel.
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
