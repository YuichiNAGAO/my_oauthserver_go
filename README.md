
# 目的

OAuth認可コードフローに対応した認可サーバーを簡易的に実装することで、OAuthへの理解を深める。

## エンドポイント


- 認可サーバーのエンドポイントについて
  - /auth: 認可エンドポイント
  - /token: トークンエンドポイント
  - /authcheck: 

## 注意点

- クライアント側の実装は行なっていない
- 認可サーバーではクライアントとユーザーをハードコーディングしている。
  - クライアント
    ```
    var clientInfo = Client{
        id:          "1234",
        name:        "test",
        redirectURL: "http://localhost:8080/callback",
        secret:      "secret",
    }
    ```
  - ユーザー
    ```
    var testUser = User{
        id:       1111,
        name:     "test",
        password: "hoge",
    }
    ```

## 動作確認の流れ

1. 認可サーバーを立ち上げる
   ```
   go run main.go
   ```
 
2. ブラウザに行き認可エンドポイントへアクセスへリクエストを送る
