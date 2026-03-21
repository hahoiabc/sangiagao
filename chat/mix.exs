defmodule RiceChat.MixProject do
  use Mix.Project

  def project do
    [
      app: :rice_chat,
      version: "0.1.0",
      elixir: "~> 1.16",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  def application do
    [
      mod: {RiceChat.Application, []},
      extra_applications: [:logger, :runtime_tools]
    ]
  end

  defp deps do
    [
      {:phoenix, "~> 1.7"},
      {:phoenix_live_dashboard, "~> 0.8"},
      {:jason, "~> 1.4"},
      {:plug_cowboy, "~> 2.7"},
      {:mongodb_driver, "~> 1.4"},
      {:joken, "~> 2.6"},
      {:cors_plug, "~> 3.0"},
      {:postgrex, "~> 0.19"}
    ]
  end
end
