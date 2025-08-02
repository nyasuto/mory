// Migration tool for generating embeddings for existing memories
// Usage:
//
//	go run migrate_embeddings.go --dry-run  # Preview migration
//	go run migrate_embeddings.go            # Execute migration
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nyasuto/mory/internal/config"
	"github.com/nyasuto/mory/internal/memory"
	"github.com/nyasuto/mory/internal/semantic"
)

func main() {
	fmt.Println("🔄 Mory 記憶マイグレーション: 既存記憶への埋め込み生成")
	fmt.Println("================================================")

	if len(os.Args) > 1 && os.Args[1] == "--dry-run" {
		fmt.Println("🔍 ドライランモード: 実際の変更は行いません")
	}

	// 設定読み込み
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Printf("設定読み込みエラー: %v", err)
		return
	}

	if cfg.Semantic == nil || !cfg.Semantic.Enabled || cfg.Semantic.OpenAIAPIKey == "" {
		log.Fatal("❌ セマンティック検索が無効です。OpenAI APIキーを設定してください。")
	}

	fmt.Printf("✅ セマンティック検索設定: 有効\n")
	fmt.Printf("   - モデル: %s\n", cfg.Semantic.EmbeddingModel)
	fmt.Printf("   - APIキー: %s...\n", cfg.Semantic.OpenAIAPIKey[:10])

	// メモリストア初期化
	var memoriesFile string
	if cfg.DataPath != "" {
		memoriesFile = cfg.DataPath
	} else {
		memoriesFile = "data/memories.json"
	}

	store := memory.NewJSONMemoryStore(memoriesFile, "data/operations.log")

	// セマンティック検索エンジン初期化
	embeddingService := semantic.NewOpenAIEmbeddingService(cfg.Semantic.OpenAIAPIKey, cfg.Semantic.EmbeddingModel)
	vectorStore := semantic.NewLocalVectorStore()
	keywordEngine := memory.NewSearchEngine(store)

	semanticSearchEngine := semantic.NewSemanticSearchEngine(
		keywordEngine,
		embeddingService,
		vectorStore,
		cfg.Semantic.HybridWeight,
		cfg.Semantic.SimilarityThreshold,
		cfg.Semantic.CacheEnabled,
	)

	// 全記憶を読み込み
	fmt.Println("\n📖 既存記憶を読み込み中...")
	memories, err := store.List("")
	if err != nil {
		log.Fatalf("記憶読み込みエラー: %v", err)
	}

	fmt.Printf("✅ %d個の記憶を発見\n", len(memories))

	// 埋め込みが必要な記憶を特定
	var needsEmbedding []*memory.Memory
	var hasEmbedding int

	for _, mem := range memories {
		if len(mem.Embedding) == 0 {
			needsEmbedding = append(needsEmbedding, mem)
		} else {
			hasEmbedding++
		}
	}

	fmt.Printf("\n📊 埋め込み状況:\n")
	fmt.Printf("   - 既に埋め込み有り: %d個\n", hasEmbedding)
	fmt.Printf("   - 埋め込み必要: %d個\n", len(needsEmbedding))

	if len(needsEmbedding) == 0 {
		fmt.Println("✅ 全ての記憶に既に埋め込みが存在します。マイグレーション不要です。")
		return
	}

	// ドライランの場合はここで終了
	if len(os.Args) > 1 && os.Args[1] == "--dry-run" {
		fmt.Println("\n🔍 ドライラン完了:")
		fmt.Printf("   - %d個の記憶に埋め込みを生成する必要があります\n", len(needsEmbedding))
		fmt.Println("   - 実際に実行するには '--dry-run' オプションを外してください")

		// サンプル表示
		fmt.Println("\n📝 埋め込み生成が必要な記憶 (最初の5個):")
		for i, mem := range needsEmbedding {
			if i >= 5 {
				break
			}
			fmt.Printf("   %d. [%s] %s: %s\n", i+1, mem.Category, mem.Key,
				truncateString(mem.Value, 40))
		}
		return
	}

	// 実際のマイグレーション実行
	fmt.Println("\n🧠 埋め込み生成を開始...")

	estimatedCost := float64(len(needsEmbedding)) * 50.0 * 0.13 / 1000000.0 // 約50トークン/記憶の想定
	fmt.Printf("   - 推定コスト: ~$%.4f (text-embedding-3-large)\n", estimatedCost)

	var successCount int
	var errorCount int

	for i, mem := range needsEmbedding {
		fmt.Printf("   [%d/%d] %s", i+1, len(needsEmbedding), mem.Key)

		err := semanticSearchEngine.GenerateEmbedding(mem)
		if err != nil {
			fmt.Printf(" ❌ エラー: %v\n", err)
			errorCount++
			continue
		}

		// 更新された記憶を保存
		_, err = store.Save(mem)
		if err != nil {
			fmt.Printf(" ❌ 保存エラー: %v\n", err)
			errorCount++
			continue
		}

		fmt.Printf(" ✅\n")
		successCount++
	}

	fmt.Printf("\n🎉 マイグレーション完了!\n")
	fmt.Printf("   - 成功: %d個\n", successCount)
	fmt.Printf("   - エラー: %d個\n", errorCount)
	fmt.Printf("   - 総記憶数: %d個\n", len(memories))

	if errorCount > 0 {
		fmt.Printf("\n⚠️  %d個の記憶で埋め込み生成に失敗しました\n", errorCount)
		fmt.Println("   - APIキーの制限またはネットワークエラーの可能性があります")
		fmt.Println("   - 後でもう一度実行することで残りの記憶を処理できます")
	}

	// 最終統計
	finalMemories, _ := store.List("")
	finalHasEmbedding := 0
	for _, mem := range finalMemories {
		if len(mem.Embedding) > 0 {
			finalHasEmbedding++
		}
	}

	coverage := float64(finalHasEmbedding) / float64(len(finalMemories)) * 100
	fmt.Printf("\n📈 最終統計:\n")
	fmt.Printf("   - 埋め込みカバレッジ: %.1f%% (%d/%d)\n",
		coverage, finalHasEmbedding, len(finalMemories))

	if coverage >= 100.0 {
		fmt.Println("✅ 全ての記憶がセマンティック検索可能になりました！")
	} else {
		fmt.Printf("⚠️  %d個の記憶がまだセマンティック検索できません\n",
			len(finalMemories)-finalHasEmbedding)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
