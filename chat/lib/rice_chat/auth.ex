defmodule RiceChat.Auth do
  @moduledoc "JWT verification - validates tokens issued by Golang backend"

  use Joken.Config

  @impl true
  def token_config do
    default_claims(skip: [:iss, :aud])
  end

  def verify_token(token) do
    secret = System.get_env("JWT_SECRET") || raise("JWT_SECRET environment variable must be set")
    signer = Joken.Signer.create("HS256", secret)

    case verify_and_validate(token, signer) do
      {:ok, claims} -> {:ok, claims["sub"]}
      {:error, reason} -> {:error, reason}
    end
  end
end
