const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    serverError: null,
    itemTypes: [],
    newItems: false,
    amazon: null,
    cashDonate: null,
    volunteer: null,
    processing: false,
  },
  mounted() {
    try {
      step = JSON.parse(localStorage.getItem('step3'));
      this.itemTypes = step.itemTypes
      this.newItems = step.newItems
      this.amazon = step.amazon
      this.cashDonate = step.cashDonate
      this.volunteer = step.volunteer
    } catch(e) {
      localStorage.removeItem('step3')
    }
  },
  methods:{
    checkForm: function (e) {
      e.preventDefault();

      this.errors = [];

      if (this.itemTypes.length == 0) {
        this.errors.push('Please select at least one item');
      }

      if (this.errors.length == 0 ) {
        this.saveStep()
        this.submitRegistration()
      }
    },
    saveStep() {
      localStorage.setItem('step3', JSON.stringify({
        itemTypes: this.itemTypes,
        newItems: this.newItems,
        amazon: this.amazon,
        cashDonate: this.cashDonate,
        volunteer: this.volunteer,
      }))
    },
    submitRegistration: function() {
      this.processing = true

      try {
        step1 = JSON.parse(localStorage.getItem('step1'))
        step2 = JSON.parse(localStorage.getItem('step2'))
      } catch(e) {
        console.log(e)
        return
      }

      org = {
        itemTypes: this.itemTypes,
        newItems: this.newItems,
        amazon: this.amazon,
        cashDonate: this.cashDonate,
        volunteer: this.volunteer,
        ...step1,
        ...step2
      }
      console.log(org)

      fetch('/api/v1/charity/register', {
        method: 'post',
        body: JSON.stringify(org)
      }).then(response => {
        console.log(response)
        if (response.ok) {
          window.location.assign("/charity/signup/thankyou");
        } else {
          this.processing = false
          if (response.status == 409) {
            this.serverError = 'That email has already been registered.'
          } else {
            this.serverError= 'Error registering. Please try again later.'
          }
          return false
        }
      })
    }
  }
})
