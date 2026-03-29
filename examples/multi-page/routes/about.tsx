import Button from "./_Button";

export const title = "About - Multi Page";

export default function About() {
  return (
    <main style={{ padding: "48px", fontFamily: "sans-serif" }}>
      <h1>About</h1>
      <p>This route lives in `routes/about.tsx`.</p>
      <Button href="/">Back home</Button>
    </main>
  );
}
