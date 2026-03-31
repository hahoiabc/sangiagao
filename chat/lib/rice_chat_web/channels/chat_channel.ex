defmodule RiceChatWeb.ChatChannel do
  use Phoenix.Channel

  alias RiceChat.Messages

  @max_message_size 4096
  @rate_limit_messages 30
  @rate_limit_window_ms 60_000
  @valid_types ~w(text image audio listing_link)

  @impl true
  def join("chat:" <> conversation_id, _params, socket) do
    user_id = socket.assigns.user_id

    case Messages.is_participant?(conversation_id, user_id) do
      true ->
        socket =
          socket
          |> assign(:conversation_id, conversation_id)
          |> assign(:msg_timestamps, [])

        {:ok, socket}

      false ->
        {:error, %{reason: "unauthorized"}}
    end
  end

  @impl true
  def handle_in("new_message", payload, socket) do
    content = payload["content"] || ""
    type = payload["type"] || "text"
    reply_to_id = payload["reply_to_id"]

    cond do
      type not in @valid_types ->
        {:reply, {:error, %{reason: "invalid message type"}}, socket}

      byte_size(content) > @max_message_size ->
        {:reply, {:error, %{reason: "message too long (max #{@max_message_size} bytes)"}}, socket}

      String.trim(content) == "" ->
        {:reply, {:error, %{reason: "message cannot be empty"}}, socket}

      rate_limited?(socket) ->
        {:reply, {:error, %{reason: "rate limited, please slow down"}}, socket}

      true ->
        socket = track_message(socket)

        message = %{
          conversation_id: socket.assigns.conversation_id,
          sender_id: socket.assigns.user_id,
          content: content,
          type: type,
          reply_to_id: reply_to_id,
          timestamp: DateTime.utc_now(),
          read_at: nil
        }

        case Messages.create(message) do
          {:ok, saved_message} ->
            broadcast!(socket, "new_message", serialize_message(saved_message))
            {:reply, :ok, socket}

          {:error, reason} ->
            {:reply, {:error, %{reason: reason}}, socket}
        end
    end
  end

  @impl true
  def handle_in("read_receipt", %{"message_id" => message_id}, socket) do
    Messages.mark_read(message_id, socket.assigns.user_id)
    broadcast!(socket, "read_receipt", %{message_id: message_id, reader_id: socket.assigns.user_id})
    {:noreply, socket}
  end

  @impl true
  def handle_in("typing", _params, socket) do
    broadcast_from!(socket, "typing", %{user_id: socket.assigns.user_id})
    {:noreply, socket}
  end

  # Relay: broadcast a message already saved by Go backend (no DB write)
  @impl true
  def handle_in("relay", payload, socket) do
    broadcast_from!(socket, "new_message", %{
      id: to_string(payload["id"] || ""),
      conversation_id: to_string(payload["conversation_id"] || socket.assigns.conversation_id),
      sender_id: to_string(payload["sender_id"] || socket.assigns.user_id),
      content: to_string(payload["content"] || ""),
      type: to_string(payload["type"] || "text"),
      reply_to_id: if(payload["reply_to_id"], do: to_string(payload["reply_to_id"]), else: nil),
      timestamp: to_string(payload["created_at"] || DateTime.to_iso8601(DateTime.utc_now())),
      read_at: nil
    })
    {:noreply, socket}
  end

  @impl true
  def handle_in("reaction", %{"message_id" => message_id, "emoji" => emoji}, socket) do
    broadcast!(socket, "reaction", %{
      message_id: message_id,
      user_id: socket.assigns.user_id,
      emoji: emoji
    })
    {:noreply, socket}
  end

  # --- Serialization ---

  defp serialize_message(msg) do
    base = %{
      id: to_string(msg[:id] || msg[:_id] || ""),
      conversation_id: to_string(msg[:conversation_id] || ""),
      sender_id: to_string(msg[:sender_id] || ""),
      content: to_string(msg[:content] || ""),
      type: to_string(msg[:type] || "text"),
      reply_to_id: if(msg[:reply_to_id], do: to_string(msg[:reply_to_id]), else: nil),
      timestamp: format_dt(msg[:timestamp]),
      read_at: format_dt(msg[:read_at])
    }
    base
  end

  defp format_dt(nil), do: nil
  defp format_dt(%DateTime{} = dt), do: DateTime.to_iso8601(dt)
  defp format_dt(other), do: to_string(other)

  # --- Rate limiting helpers ---

  defp rate_limited?(socket) do
    now = System.monotonic_time(:millisecond)
    cutoff = now - @rate_limit_window_ms
    recent = Enum.count(socket.assigns.msg_timestamps, &(&1 > cutoff))
    recent >= @rate_limit_messages
  end

  defp track_message(socket) do
    now = System.monotonic_time(:millisecond)
    cutoff = now - @rate_limit_window_ms
    timestamps = [now | Enum.filter(socket.assigns.msg_timestamps, &(&1 > cutoff))]
    assign(socket, :msg_timestamps, timestamps)
  end
end
