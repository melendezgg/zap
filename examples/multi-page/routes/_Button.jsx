export default function Button({ href, children }) {
  return (
    <a
      href={href}
      style={{
        display: "inline-block",
        padding: "10px 16px",
        borderRadius: "10px",
        background: "#111827",
        color: "#ffffff",
        textDecoration: "none",
      }}
    >
      {children}
    </a>
  );
}
