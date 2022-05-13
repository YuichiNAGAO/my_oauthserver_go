
# 目的

OAuth認可コードフローに対応した認可サーバーを簡易的に実装することで、OAuthへの理解を深める。

## 処理の流れ

- **認可サーバーのエンドポイントについて**
  - ***/auth: 認可エンドポイント***
    - <u>リクエスト [クライアント->認可サーバー]</u>
      
      *`認可コードフローでは、認可リクエストパラメーター群は、クエリー部に埋め込まれる`
      
      - パラメタ
      response_type = code (これは認可コードフローに対応する)  
      client_id  
      redirect_uri = [リダイレクトURI(これはクライアント側に実装される)]  
      state  
      scope  
    
    - <u>レスポンス [認可サーバー->ブラウザ]</u>
     
       HTML の認可ページを生成してブラウザに返す
      
  - ***/authcheck: 認可決定エンドポイント***


  - ***/token: トークンエンドポイント***
    - リクエスト [クライアント->認可サーバー]
      
      
    
      *`認可リクエストにredirect_uriリクエストパラメーターが含まれていた場合、トークンリクエストにも含まれていなければならない.`
    
      - パラメタ
      grant_type = authorization_code *テーブル1参照  
      code = [認可コード]  
      redirect_uri = [リダイレクトURI(認可リクエストに含まれていたものと同じ)]  
      
    - レスポンス

  
 

###### テーブル1: OAuthの各フローとgrant_typeの対応
| フロー | grant_type |
| :--- | :---: |
| 認可コードフロー | authorization_code |
| リソースオーナー・パスワード・クレデンシャルズフロー | password |
| クライアント・クレデンシャルズフロー | client_credentials |

## 注意点

- クライアント側の実装は行なっていない
- 認可サーバーではクライアントとユーザーの情報をハードコーディングしている。
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
   ```
   http://localhost:8080/auth?client_id=1234&response_type=code&redirect_uri=aaa
   ```
   
   
   
3. トークンエンドポイントへリクエストを送る

   ```
   CODE={発行された認可コード}
   curl http://localhost:4567/token \
   -d grant_type=authorization_code \
   -d code=$CODE -d client_id=1234 \
   -d redirect_uri=http://example.com/
   ```