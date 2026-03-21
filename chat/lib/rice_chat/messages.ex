defmodule RiceChat.Messages do
  @moduledoc "MongoDB operations for chat messages"

  @collection "messages"

  def create(message) do
    case Mongo.insert_one(:mongo, @collection, message) do
      {:ok, result} ->
        {:ok, Map.put(message, :id, result.inserted_id)}
      {:error, reason} ->
        {:error, reason}
    end
  end

  def list_by_conversation(conversation_id, opts \\ []) do
    limit = Keyword.get(opts, :limit, 50)
    before = Keyword.get(opts, :before, nil)

    filter = %{conversation_id: conversation_id}
    filter = if before, do: Map.put(filter, :timestamp, %{"$lt" => before}), else: filter

    Mongo.find(:mongo, @collection, filter,
      sort: %{timestamp: -1},
      limit: limit
    )
    |> Enum.to_list()
    |> Enum.reverse()
  end

  def mark_read(message_id, reader_id) do
    Mongo.update_one(:mongo, @collection,
      %{_id: message_id, sender_id: %{"$ne" => reader_id}},
      %{"$set" => %{read_at: DateTime.utc_now()}}
    )
  end

  def is_participant?(conversation_id, user_id) do
    query = "SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1 AND (buyer_id = $2 OR seller_id = $2))"
    case Postgrex.query(:pg, query, [uuid_to_bin(conversation_id), uuid_to_bin(user_id)]) do
      {:ok, %{rows: [[true]]}} -> true
      _ -> false
    end
  end

  defp uuid_to_bin(uuid_string) when is_binary(uuid_string) do
    uuid_string
    |> String.replace("-", "")
    |> Base.decode16!(case: :lower)
  end

  def unread_count(conversation_id, user_id) do
    Mongo.count_documents(:mongo, @collection, %{
      conversation_id: conversation_id,
      sender_id: %{"$ne" => user_id},
      read_at: nil
    })
  end
end
