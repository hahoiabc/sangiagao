defmodule RiceChatWeb.MessageController do
  use Phoenix.Controller

  alias RiceChat.Messages
  alias RiceChat.Auth

  def index(conn, %{"conversation_id" => conversation_id} = params) do
    case get_user_id(conn) do
      {:ok, _user_id} ->
        opts = []
        opts = if params["before"], do: Keyword.put(opts, :before, params["before"]), else: opts
        opts = if params["limit"], do: Keyword.put(opts, :limit, String.to_integer(params["limit"])), else: opts

        messages = Messages.list_by_conversation(conversation_id, opts)

        serialized =
          Enum.map(messages, fn msg ->
            %{
              id: to_string(msg[:_id] || msg[:id] || ""),
              conversation_id: msg[:conversation_id],
              sender_id: msg[:sender_id],
              content: msg[:content],
              type: msg[:type] || "text",
              read_at: format_datetime(msg[:read_at]),
              created_at: format_datetime(msg[:timestamp])
            }
          end)

        json(conn, %{data: serialized, total: length(serialized)})

      {:error, _} ->
        conn
        |> put_status(401)
        |> json(%{error: "unauthorized"})
    end
  end

  defp get_user_id(conn) do
    case get_req_header(conn, "authorization") do
      ["Bearer " <> token] -> Auth.verify_token(token)
      _ -> {:error, :no_token}
    end
  end

  defp format_datetime(nil), do: nil
  defp format_datetime(%DateTime{} = dt), do: DateTime.to_iso8601(dt)
  defp format_datetime(other), do: to_string(other)
end
