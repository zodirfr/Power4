document.addEventListener("DOMContentLoaded", () => {
  const cells = document.querySelectorAll(".cell");
  const vsAI = window.location.href.includes("game"); // simple check

  cells.forEach(cell => {
    cell.addEventListener("click", e => {
      if (!document.querySelector(".winner")) {
        const form = cell.closest("form");
        if (form) {
          form.submit();

          // Si mode IA, on déclenche le coup de l'IA après 1s
          if (vsAI) {
            setTimeout(() => {
              fetch("/ai-move", { method: "POST" })
                .then(() => location.reload());
            }, 1000);
          }
        }
      }
    });
  });
});

