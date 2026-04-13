export default function App() {
  return (
    <main
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        fontFamily: "sans-serif",
        background: "#f5f7fb",
        color: "#172033"
      }}
    >
      <section style={{ textAlign: "center", padding: "2rem" }}>
        <p style={{ margin: 0, fontSize: "0.875rem", letterSpacing: "0.08em" }}>
          TASKFLOW
        </p>
        <h1 style={{ marginBottom: "0.75rem" }}>Foundation Ready</h1>
        <p style={{ margin: 0, maxWidth: "32rem" }}>
          Phase 0 scaffolding is in place. Phase 1 database migrations are also
          available. Application features begin in later phases.
        </p>
      </section>
    </main>
  );
}
