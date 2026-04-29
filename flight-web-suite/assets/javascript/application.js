const dropdownToggles = document.getElementsByClassName('dropdown-toggle');
for (const dropdownToggle of dropdownToggles) {
    dropdownToggle.addEventListener("click", (e) => {
        document.getElementById(dropdownToggle.dataset.targetId).classList.toggle("hidden");
    })
}
