defmodule RiceChatWeb.UserSocket do
  use Phoenix.Socket

  channel "chat:*", RiceChatWeb.ChatChannel

  @impl true
  def connect(%{"token" => token}, socket, _connect_info) do
    case RiceChat.Auth.verify_token(token) do
      {:ok, user_id} ->
        {:ok, assign(socket, :user_id, user_id)}

      {:error, _reason} ->
        :error
    end
  end

  def connect(_params, _socket, _connect_info), do: :error

  @impl true
  def id(socket), do: "user_socket:#{socket.assigns.user_id}"
end
