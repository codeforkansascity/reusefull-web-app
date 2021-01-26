/**
 * Basic multi-step form logic.
 * Ref: https://www.w3schools.com/howto/howto_js_form_steps.asp
 */

let currentSection = 0;

const displaySection = (section) => {
  // Display active section
  const Sections = document.getElementsByClassName("form-section");
  Sections[section].style.display = "block";
  // Display correct stepper buttons
  if (section == 0) {
    document.getElementById("prevButton").style.display = "none";
  } else {
    document.getElementById("prevButton").style.display = "inline";
  }
  if (section == Sections.length - 1) {
    document.getElementById("nextButton").innerHTML = "SUBMIT";
  } else {
    document.getElementById("nextButton").innerHTML = "NEXT";
  }
};

const Step = (increment) => {
  const Sections = document.getElementsByClassName("form-section");
  Sections[currentSection].style.display = "none";
  currentSection = currentSection + increment;
  if (currentSection >= Sections.length) {
    document.getElementById("orgRegistrationForm").submit();
    return false;
  }
  displaySection(currentSection);
};

displaySection(currentSection);
