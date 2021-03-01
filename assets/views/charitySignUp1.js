const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    loaded: false,
    errors: [],
    name: null,
    contactName: null,
    email: null,
    phone: null,
    address: null,
    city: null,
    state: null,
    zip: null,
  },
  mounted() {
    this.loaded = true;
    if (localStorage.getItem('step1')) {
      try {
        step1 = JSON.parse(localStorage.getItem('step1'));
        this.name = step1.name
        this.contactName = step1.contactName
        this.email = step1.email
        this.phone = step1.phone
        this.address = step1.address
        this.city = step1.city
        this.state = step1.state
        this.zip = step1.zip
      } catch(e) {
        localStorage.removeItem('step1')
      }
    }
  },
  methods:{
    saveStep() {
      localStorage.setItem('step1', JSON.stringify({
        name: this.name,
        contactName: this.contactName,
        email: this.email,
        // filter out non-num chars on submission
        phone: this.phone.replace(/[^\d]/g, ""),
        address: this.address,
        city: this.city,
        state: this.state,
        zip: this.zip
      }))
    },
    checkForm: function (e) {
      e.preventDefault();

      if (this.name &&
          this.contactName &&
          this.email &&
          this.phone &&
          this.address &&
          this.city &&
          this.state && this.state != "Choose..." &&
          this.zip) {
        this.saveStep()
        window.location.assign("/charity/signup/step/2")
        return true;
      }

      this.errors = [];

      if (!this.name) {
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
      if (this.state == "Choose...") {
        this.errors.push('State required.');
      }
      if (!this.zip) {
        this.errors.push('Zip required.');
      }

      console.log(this.errors.length)
    }
  },
  watch: {
      phone: function(num) {
        /*
            On change, filter any non-nums, and 
            return formatted phone num for display.
        */
        const clean = num.replace(/[^\d]/g, "")
        const cleanLen = clean.length

        if (cleanLen < 4) {
            this.phone = clean;
        } else if (cleanLen < 7) {
            this.phone = `(${clean.slice(0, 3)}) ${clean.slice(3)}`
        } else {
            this.phone = `(${clean.slice(0,3)}) ${clean.slice(3, 6)}-${clean.slice(6, 10)}`
        }
        
        return this.phone
    }
  }
})
