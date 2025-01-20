/**
 * @param {HTMLLIElement} element
 */
function dropdownSelectItem(element) {
  const name = element.innerText;
  const val = element.getAttribute("data-value");
  const dropdownID = element.getAttribute("data-for");

  document.getElementById(`${dropdownID}-selected-value`).innerText = name;
  document.getElementById(`${dropdownID}-input`).value = val;
  document
    .getElementById(dropdownID)
    .querySelector("ul")
    .classList.add("hidden");

  document
    .getElementById(dropdownID)
    .querySelector("button")
    .classList.add("rounded-b-md");
}

var dropdowns = [];
window.addEventListener("load", () => {
  setInterval(() => {
    const allDropdowns = document.querySelectorAll("div.dropdown");
    let newDropdowns = [];
    allDropdowns.forEach((elem) => {
      if (dropdowns.every((oldD) => oldD.id != elem.id)) {
        newDropdowns.push(elem);
        dropdowns.push(elem);
      }
    });
    for (const dropdown of newDropdowns) {
      dropdown.querySelector("button").addEventListener("click", (e) => {
        const options = dropdown.querySelector("ul");

        if (options) {
          if (options.classList.contains("hidden")) {
            options.classList.remove("hidden");
            e.target.classList.remove("rounded-b-md");
          } else {
            options.classList.add("hidden");
            e.target.classList.add("rounded-b-md");
          }
        }
      });

      const options = dropdown.querySelectorAll("ul>li");
      for (const option of options) {
        option.addEventListener("click", () => dropdownSelectItem(option));
      }
    }
  }, 400);
});
