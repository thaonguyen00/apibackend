package agentDAO

import (
	"context"
	klog "github.com/go-kit/kit/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strings"
)

type Collection interface {
	Find(ctx context.Context, request Request) (*mongo.Cursor, error)
}

type AgentCollection struct {
	DB *mongo.Collection

}

func NewCollection(DB *mongo.Collection) Collection {
	var agentCol Collection
	agentCol = &AgentCollection{DB}

	logger := klog.NewLogfmtLogger(os.Stderr)
	agentCol = &LoggingMW{logger, agentCol}

	agentCol = &TracingMW{agentCol}
	return agentCol

}

type Request struct {
	Filter           map[string]interface{}
	ProjectionFields string //string rep of a list of fields, will convert to bson.m list
	Offset           int64
	Limit            int64
	SortKey          string
	SortOrder        int64
}

func (m *AgentCollection) Find(ctx context.Context, request Request) (*mongo.Cursor, error){

	projection := createProjection(request.ProjectionFields)
	cursor, err := m.DB.Find(
		ctx,
		request.Filter,
		options.Find().SetProjection(projection).SetLimit(int64(request.Limit)).SetSkip(int64(request.Offset)).SetSort(bson.D{{request.SortKey, request.SortOrder}}),
	)

	return cursor, err
}

//======================= helper functions =======================
func createProjection(fields string) *bson.M {
	projectionFields := []string{}

	//convert from string to a list of string
	if fields != "" {
		projectionFields= strings.Split(strings.Replace(fields, " ", "", -1), ",")
	}

	projection := bson.M{}
	for _, fields := range projectionFields {
		projection[fields] = 1
	}
	return &projection
}
