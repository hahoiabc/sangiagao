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
          |> assign(:in_call, false)

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
    socket = assign(socket, :in_call, true)

    # Auto-timeout if not answered
    Process.send_after(self(), :call_timeout, @call_timeout_ms)

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
    socket = assign(socket, :in_call, true)
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
    socket = assign(socket, :in_call, false)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_reject", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast!(socket, "call_reject", payload)
    socket = assign(socket, :in_call, false)
    {:noreply, socket}
  end

  @impl true
  def handle_in("call_busy", _params, socket) do
    payload = %{user_id: socket.assigns.user_id}
    broadcast!(socket, "call_busy", payload)
    {:noreply, socket}
  end

  @impl true
  def handle_info(:call_timeout, socket) do
    if socket.assigns.in_call do
      {:noreply, socket}
    else
      broadcast!(socket, "call_timeout", %{})
      {:noreply, socket}
    end
  end
end
