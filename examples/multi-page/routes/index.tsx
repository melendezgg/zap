import Button from "./_Button";

export const title = "Home - Multi Page";

export default function App() {
  return (
    <main style={{ padding: "48px", fontFamily: "sans-serif" }}>
      <h1>Multi-page example</h1>
      <p>This app uses multiple static routes.</p>
      <Button href="/about">Go to about</Button>
    </main>
  );
}
