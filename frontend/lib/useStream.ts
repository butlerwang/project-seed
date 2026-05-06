import { useCallback, useRef, useState } from "react";

import { getToken } from "./auth";

interface StreamOptions {
  operation?: string;
  system?: string;
}

interface StreamState {
  text: string;
  loading: boolean;
  error: string | null;
}

/**
 * Usage in a component:
 * const { text, loading, error, stream, abort } = useStream({ operation: "research" });
 * <button onClick={() => stream("Summarize the latest trends in AI")}>Ask</button>
 * <pre>{text}</pre>
 */
export function useStream(opts: StreamOptions = {}) {
  const [state, setState] = useState<StreamState>({ text: "", loading: false, error: null });
  const abortRef = useRef<AbortController | null>(null);

  const abort = useCallback(() => {
    abortRef.current?.abort();
    setState((current) => ({ ...current, loading: false }));
  }, []);

  const stream = useCallback(
    async (prompt: string) => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setState({ text: "", loading: true, error: null });

      try {
        const base = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
        const resp = await fetch(`${base}/api/v1/llm/stream`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${getToken() ?? ""}`,
          },
          body: JSON.stringify({
            operation: opts.operation ?? "chat",
            system: opts.system ?? "",
            prompt,
          }),
          signal: controller.signal,
        });

        if (!resp.ok) {
          const msg = await resp.text();
          setState({ text: "", loading: false, error: msg || "Stream failed" });
          return;
        }

        const reader = resp.body?.getReader();
        if (!reader) {
          setState({ text: "", loading: false, error: "No response body" });
          return;
        }

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() ?? "";

          for (const line of lines) {
            if (!line.startsWith("data: ")) continue;
            const token = line.slice(6);
            if (token === "[DONE]") {
              setState((current) => ({ ...current, loading: false }));
              return;
            }
            if (token === "[ERROR]") {
              setState((current) => ({ ...current, loading: false, error: "LLM error" }));
              return;
            }
            setState((current) => ({ ...current, text: current.text + token }));
          }
        }

        setState((current) => ({ ...current, loading: false }));
      } catch (err: unknown) {
        if (err instanceof Error && err.name === "AbortError") {
          return;
        }
        setState({
          text: "",
          loading: false,
          error: err instanceof Error ? err.message : "Unknown error",
        });
      }
    },
    [opts.operation, opts.system]
  );

  return { ...state, stream, abort };
}
