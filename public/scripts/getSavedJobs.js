document.addEventListener("DOMContentLoaded", () => {
  htmx
    .ajax("GET", "/jobs/get/saved", { target: "#container", swap: "beforeend" })
    .then(() => {
      document.body.classList.remove("htmx-request")
    });
});
