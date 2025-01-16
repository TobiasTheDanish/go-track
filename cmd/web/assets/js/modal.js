window.addEventListener("load", () => {
  const dismissBtn = document.getElementById("modal-dismiss-btn");

  if (dismissBtn) {
    dismissBtn.addEventListener("click", () => {
      const container = document.getElementById("modal-container");
      container.classList.add("hidden");
    });
  }
});
