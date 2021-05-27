const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    website: null,
    budget: null,
    dropoff: null,
    pickup: null,
    faith: null,
    resell: null,
    taxID: null,
    logo: null,
    charityTypes: [],
    other: null,
    mission: null,
    description: null
  },
  mounted() {
    if (localStorage.getItem('step2')) {
      try {
        step = JSON.parse(localStorage.getItem('step2'));
        this.website = step.website
        this.budget = step.budget
        this.dropoff = step.dropoff
        this.pickup = step.pickup
        this.faith = step.faith
        this.resell = step.resell
        this.taxID = step.taxID
        this.logo = step.logo
        this.charityTypes = step.charityTypes
        this.other = step.other
        this.mission = step.mission
        this.description = step.description
      } catch(e) {
        localStorage.removeItem('step2')
      }
    }
  },
  methods:{
    filePicked(event) {
      if (event.target.files.length == 0 ) {
        console.log("no file picked")
        return
      }
      file = event.target.files[0]
      reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = e =>{
        this.logo = e.target.result;
        console.log(this.logo)
      }

    },
    saveStep() {
      localStorage.setItem('step2', JSON.stringify({
        website: this.website,
        budget: this.budget,
        dropoff: this.dropoff,
        pickup: this.pickup,
        faith: this.faith,
        resell: this.resell,
        taxID: this.taxID,
        logo: this.logo,
        charityTypes: this.charityTypes,
        other: this.other,
        mission: this.mission,
        description: this.description
      }))
    },
    checkForm: function (e) {
      e.preventDefault();

      this.errors = [];

      if (!this.website) {
        this.errors.push('Website required.');
      }
      if (!this.budget) {
        this.errors.push('Budget required.');
      }
      if (!this.pickup && !this.dropoff) {
        this.errors.push('Pick-up or Drop-off option required.')
      }
      if (!this.taxID) {
        this.errors.push('Tax ID required.');
      }
      if (this.charityTypes.length == 0 && !this.other) {
        this.errors.push('At least one charity type is required')
      }
      if (!this.mission) {
        this.errors.push('Mission required.');
      }
      if (!this.description) {
        this.errors.push('Description required.');
      }

      if (this.errors.length == 0) {
        this.saveStep();
        window.location.assign("/charity/signup/step/3")
        return true;
      }
    }
  }
})
