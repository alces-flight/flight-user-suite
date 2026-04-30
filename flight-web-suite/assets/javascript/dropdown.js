const dropdownToggles = document.getElementsByClassName('dropdown-toggle');
for (const dropdownToggle of dropdownToggles) {
  const dropdown = document.getElementById(dropdownToggle.dataset.targetId);
  dropdownToggle.addEventListener("click", () => {
    dropdown.classList.toggle("hidden");
  });
  addEventListener("click", (event) => {
    if (!dropdown.contains(event.target) && !dropdownToggle.contains(event.target)) {
      dropdown.classList.add("hidden");
    }
  });
}
