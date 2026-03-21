defmodule RiceChatWeb.HealthController do
  use Phoenix.Controller

  def index(conn, _params) do
    json(conn, %{status: "ok", service: "rice_chat"})
  end
end
