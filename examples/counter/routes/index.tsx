import { useState } from "react";

export const title = "Counter - Zap";

export default function App() {
  const [count, setCount] = useState(0);

  return (
    <main style={{ padding: "48px", fontFamily: "sans-serif" }}>
      <h1>Counter</h1>
      <p>Current value: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </main>
  );
}
