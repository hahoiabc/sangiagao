defmodule RiceChatWeb.Router do
  use Phoenix.Router

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/api", RiceChatWeb do
    pipe_through :api

    get "/health", HealthController, :index
    get "/conversations/:conversation_id/messages", MessageController, :index
  end
end
