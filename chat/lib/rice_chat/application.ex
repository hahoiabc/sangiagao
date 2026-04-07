defmodule RiceChat.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      # PubSub for Phoenix channels
      {Phoenix.PubSub, name: RiceChat.PubSub},
      # MongoDB connection
      {Mongo, [
        name: :mongo,
        url: System.get_env("MONGO_URL", "mongodb://localhost:27017/rice_chat"),
        pool_size: 30
      ]},
      # PostgreSQL connection (conversations live in PG)
      {Postgrex, [
        name: :pg,
        hostname: System.get_env("PG_HOST") || raise("PG_HOST must be set"),
        port: String.to_integer(System.get_env("PG_PORT", "5432")),
        username: System.get_env("PG_USER") || raise("PG_USER must be set"),
        password: System.get_env("PG_PASSWORD") || raise("PG_PASSWORD must be set"),
        database: System.get_env("PG_DATABASE") || raise("PG_DATABASE must be set")
      ]},
      # Phoenix endpoint
      RiceChatWeb.Endpoint
    ]

    opts = [strategy: :one_for_one, name: RiceChat.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
