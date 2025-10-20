document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".cell").forEach(cell => {
    cell.addEventListener("click", () => {
      if (!document.querySelector(".winner")) {
        const col = cell.dataset.col;
        window.location.href = `/play?col=${col}`;
      }
    });
  });
});