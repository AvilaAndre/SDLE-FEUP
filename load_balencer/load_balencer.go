package load_balencer

import (
	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
	"sdle.com/mod/hash_ring"
	"sdle.com/mod/database_node"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	W.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "pong"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func getHealthyNodesForID(listId string) []*hash_ring.NodeInfo {
	healthyNodes := ring.GetHealthyNodesForID(listId)

	var healthyNodesStack utils.Stack

	// Scrambles N first healthy replicas so a quorum can be performed for this key
	rand.Shuffle(min(len(healthyNodes), replicationFactor), func(i, j int) { healthyNodes[i], healthyNodes[j] = healthyNodes[j], healthyNodes[i] })

}