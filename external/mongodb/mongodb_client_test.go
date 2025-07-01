/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/test"
	"github.com/rulego/rulego/test/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDBClient(t *testing.T) {
	// 如果设置了跳过 MongoDB 测试，则跳过
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB tests")
	}

	// 检查是否有可用的 MongoDB 服务器
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	// 测试 MongoDB 连接可用性
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skipf("MongoDB server not available: %v", err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skipf("MongoDB server not responding: %v", err)
	}

	Registry := &types.SafeComponentSlice{}
	Registry.Add(&ClientNode{})
	var targetNodeType = "x/mongodbClient"

	t.Run("InitNode", func(t *testing.T) {
		node, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server":     mongoURL,
			"database":   "test_db",
			"collection": "test_collection",
			"opType":     "insert",
		}, Registry)
		assert.Nil(t, err)
		assert.NotNil(t, node)

		clientNode := node.(*ClientNode)
		assert.Equal(t, mongoURL, clientNode.Config.Server)
		assert.Equal(t, "test_db", clientNode.Config.Database)
		assert.Equal(t, "test_collection", clientNode.Config.Collection)
		assert.Equal(t, "insert", clientNode.Config.OpType)
	})

	t.Run("InsertDocument", func(t *testing.T) {
		node, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server":     mongoURL,
			"database":   "test_db",
			"collection": "test_collection",
			"opType":     "insert",
		}, Registry)
		if err != nil {
			t.Skipf("Failed to create MongoDB node: %v", err)
		}

		config := types.NewConfig()
		ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err error) {
			if err != nil {
				t.Logf("MongoDB operation result: %s, error: %v", relationType, err)
			} else {
				assert.Equal(t, types.Success, relationType)
			}
		})

		metaData := types.NewMetadata()
		metaData.PutValue("operation", "insert")

		msg := ctx.NewMsg("TEST_MSG_TYPE", metaData, "{\"name\":\"test\",\"value\":123}")

		clientNode := node.(*ClientNode)
		clientNode.OnMsg(ctx, msg)

		// 等待消息处理
		time.Sleep(time.Millisecond * 200)

		clientNode.Destroy()
	})

	t.Run("FindDocument", func(t *testing.T) {
		node, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server":     mongoURL,
			"database":   "test_db",
			"collection": "test_collection",
			"opType":     "find",
		}, Registry)
		if err != nil {
			t.Skipf("Failed to create MongoDB node: %v", err)
		}

		config := types.NewConfig()
		ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string, err error) {
			if err != nil {
				t.Logf("MongoDB find operation result: %s, error: %v", relationType, err)
			} else {
				assert.Equal(t, types.Success, relationType)
			}
		})

		metaData := types.NewMetadata()
		metaData.PutValue("operation", "find")

		msg := ctx.NewMsg("TEST_MSG_TYPE", metaData, "{\"name\":\"test\"}")

		clientNode := node.(*ClientNode)
		clientNode.OnMsg(ctx, msg)

		// 等待消息处理
		time.Sleep(time.Millisecond * 200)

		clientNode.Destroy()
	})
}

func TestMongoDBClientConfig(t *testing.T) {
	// 如果设置了跳过 MongoDB 测试，则跳过
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB tests")
	}

	Registry := &types.SafeComponentSlice{}
	Registry.Add(&ClientNode{})
	var targetNodeType = "x/mongodbClient"

	t.Run("EmptyURIConfig", func(t *testing.T) {
		_, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server": "",
		}, Registry)
		assert.NotNil(t, err)
	})

	t.Run("InvalidURIConfig", func(t *testing.T) {
		_, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server":     "invalid://localhost:27017",
			"database":   "test_db",
			"collection": "test_collection",
			"opType":     "insert",
		}, Registry)
		// 应该能创建节点，但连接会失败
		assert.Nil(t, err)
	})

	t.Run("MissingDatabaseConfig", func(t *testing.T) {
		_, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"uri":        "mongodb://localhost:27017",
			"collection": "test_collection",
			"operation":  "insert",
		}, Registry)
		assert.NotNil(t, err)
	})

	t.Run("MissingCollectionConfig", func(t *testing.T) {
		_, err := test.CreateAndInitNode(targetNodeType, types.Configuration{
			"server":   "mongodb://localhost:27017",
			"database": "test_db",
			"opType":   "insert",
		}, Registry)
		assert.NotNil(t, err)
	})
}
