{ pkgs, ... }: {
  channel = "unstable";

  packages = [
    pkgs.go
    pkgs.gcc
    pkgs.nodejs_20
    pkgs.nodePackages.nodemon
    pkgs.postgresql_15
  ];

  env = {
    # 2. 统一 DSN 格式
    DB_DSN = "host=127.0.0.1 user=postgres dbname=football_db sslmode=disable";
    GIN_MODE = "debug";
    JWT_SECRET = "a-secure-secret-for-local-idx";
    PORT = "8080";
  };

  idx = {
    extensions = [
      "golang.go"
      "google.gemini-cli-vscode-ide-companion"
      "cweijan.vscode-postgresql-client2"
    ];

    workspace = {
      onCreate = {
        # 3. 自动初始化数据库逻辑
        setup-db = ''
          # 循环检查 Postgres 是否就绪
          echo "Waiting for PostgreSQL to be ready..."
          until pg_isready -h 127.0.0.1; do
            sleep 1
          done

          # 创建数据库 (如果不存在)
          # 注意: IDX 环境中默认用户可能是 'postgres' 或当前用户名
          psql -h 127.0.0.1 -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'football_db'" | grep -q 1 || \
          psql -h 127.0.0.1 -U postgres -c "CREATE DATABASE football_db"

          # 运行初始化脚本
          echo "Running init.sql..."
          psql -h 127.0.0.1 -U postgres -d football_db -f backend/sql/init.sql
          
          echo "Database initialization completed."
        '';
        default.openFiles = [ "backend/cmd/main.go" "backend/sql/init.sql" ];
      };
    };

    previews = {
      enable = true;
      previews = {
        web = {
          command = [
            "nodemon"
            "--signal" "SIGHUP"
            "-w" "."
            "-e" "go,html"
            "-x" "cd backend && go run cmd/main.go"
          ];
          manager = "web";
        };
      };
    };
  };
}
