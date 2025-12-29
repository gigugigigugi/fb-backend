{ pkgs, ... }: {
  channel = "unstable";

  # 1. Add docker-compose to the environment's packages
  packages = [
    pkgs.go
    pkgs.gcc
    pkgs.nodejs_20
    pkgs.nodePackages.nodemon
    pkgs.docker
    pkgs.docker-compose # Makes the `docker-compose` command available
  ];

  # 3. Set the environment variables your Go application needs
  env = {
    # This DSN is for your Go app running in the preview (via nodemon).
    # It connects to the port that the Docker container exposes to `localhost`.
    DB_DSN = "postgres://user:password@localhost:5432/football_db?sslmode=disable";
    GIN_MODE = "debug";
    JWT_SECRET = "a-secure-secret-for-testing-from-compose";
  };

  idx = {
    extensions = [
      "golang.go"
      "google.gemini-cli-vscode-ide-companion"
      "ms-azuretools.vscode-docker"
    ];

    workspace = {
      # We are removing the `onStart` hook to simplify things.
      # You will run the docker-compose command manually.
      onCreate = {
        default.openFiles = ["backend/cmd/main.go" "docker-compose.yml"];
      };
    };

    # Your original preview configuration is kept exactly as it was.
    previews = {
      enable = true;
      previews = {
        web = {
          command = [
            "nodemon"
            "--signal"
            "SIGHUP"
            "-w"
            "."
            "-e"
            "go,html"
            "-x"
            "cd backend && go run cmd/main.go -addr localhost:$PORT"
          ];
          manager = "web";
        };
      };
    };
  };
}
