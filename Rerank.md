# Rerank API文档

**简介**:Rerank API文档

## 接入Dify
模型供应商选择Jina，按要求填写模型信息即可接入Dify。

## 请求方式

Post: /v1/rerank

Request:

```json
{
  "model": "jina-reranker-v2-base-multilingual",
  "query": "What is the capital of the United States?",
  "top_n": 3,
  "documents": [
    "Carson City is the capital city of the American state of Nevada.",
    "The Commonwealth of the Northern Mariana Islands is a group of islands in the Pacific Ocean. Its capital is Saipan.",
    "Washington, D.C. (also known as simply Washington or D.C., and officially as the District of Columbia) is the capital of the United States. It is a federal district.",
    "Capitalization or capitalisation in English grammar is the use of a capital letter at the start of a word. English usage varies from capitalization in other languages.",
    "Capital punishment (the death penalty) has existed in the United States since beforethe United States was a country. As of 2017, capital punishment is legal in 30 of the 50 states."
  ]
}
```

Response:

```json
{
  "results": [
    {
      "document": {
        "text": "Washington, D.C. (also known as simply Washington or D.C., and officially as the District of Columbia) is the capital of the United States. It is a federal district."
      },
      "index": 2,
      "relevance_score": 0.9999702
    },
    {
      "document": {
        "text": "Carson City is the capital city of the American state of Nevada."
      },
      "index": 0,
      "relevance_score": 0.67800725
    },
    {
      "document": {
        "text": "Capitalization or capitalisation in English grammar is the use of a capital letter at the start of a word. English usage varies from capitalization in other languages."
      },
      "index": 3,
      "relevance_score": 0.02800752
    }
  ],
  "usage": {
    "prompt_tokens": 158,
    "completion_tokens": 0,
    "total_tokens": 158
  }
}
```