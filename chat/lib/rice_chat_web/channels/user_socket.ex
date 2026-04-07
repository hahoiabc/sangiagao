defmodule RiceChatWeb.UserSocket do
  use Phoenix.Socket

  channel "chat:*", RiceChatWeb.ChatChannel

  @impl true
  def connect(params, socket, connect_info) do
    # Prefer token from Authorization header, fallback to query param
    token =
      case get_header_token(connect_info) do
        {:ok, t} -> t
        :error -> params["token"]
      end

    if token do
      case RiceChat.Auth.verify_token(token) do
        {:ok, user_id} ->
          {:ok, assign(socket, :user_id, user_id)}

        {:error, _reason} ->
          :error
      end
    else
      :error
    end
  end

  defp get_header_token(%{x_headers: headers}) do
    case List.keyfind(headers, "x-auth-token", 0) do
      {_, token} -> {:ok, token}
      nil -> :error
    end
  end

  defp get_header_token(_), do: :error

  @impl true
  def id(socket), do: "user_socket:#{socket.assigns.user_id}"
end
