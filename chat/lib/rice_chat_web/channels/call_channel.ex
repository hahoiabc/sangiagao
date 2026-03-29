defmodule RiceChatWeb.CallChannel do
  use Phoenix.Channel

  alias RiceChat.Messages

  @call_timeout_ms 60_000

  @impl true
  def join("call:" <> conversation_id, _params, socket) do
    user_id = socket.assigns.user_id

    case Messages.is_participant?(conversation_id, user_id) do
      true ->
        socket =
          socket
          |> assign(:conversation_id, conversation_id)
          |> assign(:call_answered, false)
          |> assign(:timeout_ref, nil)

        {:ok, socket}

      false ->
        {:error, %{reason: "unauthorized"}}
    end
  end

  @impl true
  def handle_in("call_initiate", %{"callee_id" => callee_id, "call_type" => call_type}, socket)
      when call_type in ~w(audio video) do
    caller_id = socket.assigns.user_id
    conversation_id = socket.assigns.conversation_id

    payload = %{
      caller_id: caller_id,
      callee_id: callee_id,
      call_type: call_type,
      conversation_id: conversation_id
    }

    broadcast!(socket, "call_initiate", payload)

    # Auto-timeout if not answered within 60s
    ref = Process.send_after(self(), :call_timeout, @call_timeout_ms)
    socket = assign(socket, :timeout_ref, ref)

    {:reply, :ok, socket}
  end

  @impl true
  def handle_in("call_offer", %{"sdp" => sdp}, socket) do
    payload = %{
      sdp: sdp,
      caller_id: socket.assigns.user_id
    }

    broadcast_from!(socket, "call_offer", payload)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_answer", %{"sdp" => sdp}, socket) do
    payload = %{
      sdp: sdp,
      callee_id: socket.assigns.user_id
    }

    broadcast_from!(socket, "call_answer", payload)

    # Cancel timeout — call was answered
    cancel_timeout(socket.assigns.timeout_ref)
    socket =
      socket
      |> assign(:call_answered, true)
      |> assign(:timeout_ref, nil)

    {:noreply, socket}
  end

  @impl true
  def handle_in("ice_candidate", %{"candidate" => candidate}, socket) do
    payload = %{
      candidate: candidate,
      from: socket.assigns.user_id
    }

    broadcast_from!(socket, "ice_candidate", payload)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_end", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast!(socket, "call_end", payload)
    cancel_timeout(socket.assigns.timeout_ref)
    socket = assign(socket, :timeout_ref, nil)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_reject", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast!(socket, "call_reject", payload)
    cancel_timeout(socket.assigns.timeout_ref)
    socket = assign(socket, :timeout_ref, nil)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_busy", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast!(socket, "call_busy", payload)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_ready", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast_from!(socket, "call_ready", payload)
    {:noreply, socket}
  end

  @impl true
  def handle_info(:call_timeout, socket) do
    unless socket.assigns.call_answered do
      broadcast!(socket, "call_timeout", %{})
    end

    {:noreply, socket}
  end

  # Cleanup when user disconnects
  @impl true
  def terminate(_reason, socket) do
    cancel_timeout(socket.assigns[:timeout_ref])

    unless socket.assigns[:call_answered] do
      broadcast!(socket, "call_end", %{user_id: socket.assigns.user_id, reason: "disconnected"})
    end

    :ok
  end

  defp cancel_timeout(nil), do: :ok
  defp cancel_timeout(ref), do: Process.cancel_timer(ref)
end
