package language

import (
	"encoding/json"
	"os"
	"path"
	"sort"
)

type Information struct {
	Name  string
	Size  uint
	Color string
}

type Statistics map[string]uint

func CreateInformationsFromStatistics(statistics Statistics) []Information {
	data, err := os.ReadFile(path.Join(*ConfigDir, "language_colors.json"))
	if err != nil {
		panic(err)
	}

	colors := make(map[string]string)

	err = json.Unmarshal(data, &colors)
	if err != nil {
		panic(err)
	}

	var infos []Information

	for k, v := range statistics {
		infos = append(infos, Information{k, v, colors[k]})
	}

	sort.Slice(infos, func(i int, j int) bool {
		return infos[j].Size < infos[i].Size
	})

	return infos
}
