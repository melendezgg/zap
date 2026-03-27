document.addEventListener("DOMContentLoaded", () => {
  const note = document.createElement("p");
  note.textContent = "Static JavaScript is served from routes/script.js.";
  note.className = "script-note";

  const main = document.querySelector("main");
  if (main) {
    main.appendChild(note);
    return;
  }

  document.body.appendChild(note);
});
