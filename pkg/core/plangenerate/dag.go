package plangenerate

import (
	"github.com/heimdalr/dag"
	"github.com/rs/zerolog/log"
)

// SortSteps sorts the steps using the Topological Sort Algorithm
func SortSteps(steps []Step) ([]Step, error) {
	d := dag.NewDAG()
	stepMap := make(map[string]Step)
	vertexIds := make(map[string]string)

	// add vertices
	for _, step := range steps {
		err := d.AddVertexByID(step.ID, step.Name)
		if err != nil {
			return nil, err
		}

		log.Trace().Str("step", step.Name).Str("id", step.ID).Msg("adding step to dag")
		stepMap[step.ID] = step
		vertexIds[step.Name] = step.ID
	}

	// add edges
	for _, step := range steps {
		for _, dep := range step.RunAfter {
			log.Trace().Str("from", vertexIds[dep]).Str("from_name", dep).Str("to", step.ID).Str("to_name", step.Name).Msg("adding dep for step")
			if err := d.AddEdge(vertexIds[dep], step.ID); err != nil {
				return nil, err
			}
		}
	}

	// visit the graph
	iv := &idVisitor{}
	d.OrderedWalk(iv)

	// log result
	log.Debug().Strs("sorted", iv.IDs).Msg("topological sort result: " + d.String())

	// prepare result
	sortedSteps := make([]Step, 0, len(iv.IDs))
	for i, id := range iv.IDs {
		s := stepMap[id]
		s.Order = i
		sortedSteps = append(sortedSteps, s)
	}

	return sortedSteps, nil
}

type idVisitor struct {
	IDs []string
}

func (pv *idVisitor) Visit(v dag.Vertexer) {
	id, _ := v.Vertex()
	pv.IDs = append(pv.IDs, id)
}
