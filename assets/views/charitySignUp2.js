const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    loaded: false,
    errors: [],
    website: null,
    budget: null,
    dropoff: null,
    pickup: null,
    faith: null,
    taxID: null,
    logo: null,
    zip: null,
    charityType: [],
    mission: null,
    description: null
  },
  mounted() {
    this.loaded = true
    if (localStorage.getItem('step2')) {
      try {
        step1 = JSON.parse(localStorage.getItem('step2'));
        // this.orgName = step1.orgName
        // this.contactName = step1.contactName
        // this.email = step1.email
        // this.phone = step1.phone
        // this.address = step1.address
        // this.city = step1.city
        // this.state = step1.state
        // this.zip = step1.zip
      } catch(e) {
        localStorage.removeItem('step2')
      }
    }
  },
  methods:{
    saveStep() {
      localStorage.setItem('step2', JSON.stringify({
        // orgName: this.orgName,
        // contactName: this.contactName,
        // email: this.email,
        // phone: this.phone,
        // address: this.address,
        // city: this.city,
        // state: this.state,
        // zip: this.zip
      }))
    },
    checkForm: function (e) {
      e.preventDefault();
      //
      // if (this.orgName &&
      //     this.contactName &&
      //     this.email &&
      //     this.phone &&
      //     this.address &&
      //     this.city &&
      //     this.state && this.state != "Choose..." &&
      //     this.zip) {
      //   this.errors = [];
      //   this.saveStep()
      //   window.location.assign("/charity/signup/step/3")
      //   return true;
      // }
      //
      // this.errors = [];
      //
      // if (!this.orgName) {
      //   this.errors.push('Organization Name required.');
      // }
      // if (!this.contactName) {
      //   this.errors.push('Contact Name required.');
      // }
      // if (!this.email) {
      //   this.errors.push('Email required.');
      // }
      // if (!this.phone) {
      //   this.errors.push('Phone required.');
      // }
      // if (!this.address) {
      //   this.errors.push('Address required.');
      // }
      // if (!this.city) {
      //   this.errors.push('City required.');
      // }
      // if (this.state == "Choose...") {
      //   this.errors.push('State required.');
      // }
      // if (!this.zip) {
      //   this.errors.push('Zip required.');
      // }
    }
  }
})
