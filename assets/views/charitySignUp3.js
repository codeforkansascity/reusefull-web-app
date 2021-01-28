const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    orgName: null,
    contactName: null,
    email: null,
    phone: null,
    address: null,
    city: null,
    state: null,
    zip: null
  },
  methods:{
    checkForm: function (e) {
      return true;
      if (this.orgName &&
          this.contactName &&
          this.email &&
          this.phone &&
          this.address &&
          this.city &&
          this.state &&
          this.zip) {
        return true;
      }

      this.errors = [];

      if (!this.orgName) {
        this.errors.push('Organization Name required.');
      }
      if (!this.contactName) {
        this.errors.push('Contact Name required.');
      }
      if (!this.email) {
        this.errors.push('Email required.');
      }
      if (!this.phone) {
        this.errors.push('Phone required.');
      }
      if (!this.address) {
        this.errors.push('Address required.');
      }
      if (!this.city) {
        this.errors.push('City required.');
      }
      if (!this.state) {
        this.errors.push('State required.');
      }
      if (!this.zip) {
        this.errors.push('Zip required.');
      }

      e.preventDefault();
    }
  }
})
