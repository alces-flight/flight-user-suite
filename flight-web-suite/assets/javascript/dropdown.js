const dropdownToggles = document.getElementsByClassName('dropdown-toggle');
for (const dropdownToggle of dropdownToggles) {
  dropdownToggle.addEventListener("click", () => {
    document.getElementById(dropdownToggle.dataset.targetId).classList.toggle("hidden");
  });
}
