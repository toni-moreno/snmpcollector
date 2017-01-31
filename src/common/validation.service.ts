export class ValidationService {
    static getValidatorErrorMessage(validatorName: string, validatorValue?: any) {
        let config = {
            'required': 'Required',
            'invalidCreditCard': 'Is invalid credit card number',
            'invalidEmailAddress': 'Invalid email address',
            'invalidPassword': 'Invalid password. Password must be at least 6 characters long, and contain a number.',
            'invalidOID': 'Invalid OID. OID must start w/ a dot and finish w/o a dot',
            'invalidFloat': 'Invalid number. Must be a float. i.e: 3.2, -4',
            'invalidInteger': 'Invalid number. Must be an integer',
            'invalidExtraTags': 'Invalid format. Must be key=value, separated by commas',
            'invalidWhiteSpaces': 'Invalid. Can\'t contain white spaces'

        };

        return config[validatorName];
    }

    static emptySelector(control) {
        if (control.value) {
            if (control.value.length === 0) return {'required' : true };
        }
         return null;
    }

    static creditCardValidator(control) {
        // Visa, MasterCard, American Express, Diners Club, Discover, JCB
        if (control.value.match(/^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\d{3})\d{11})$/)) {
            return null;
        } else {
            return { 'invalidCreditCard': true };
        }
    }

    static emailValidator(control) {
        // RFC 2822 compliant regex
        if (control.value.match(/[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?/)) {
            return null;
        } else {
            return { 'invalidEmailAddress': true };
        }
    }

    static passwordValidator(control) {
        // {6,100}           - Assert password is between 6 and 100 characters
        // (?=.*[0-9])       - Assert a string has at least one number
        if (control.value.match(/^(?=.*[0-9])[a-zA-Z0-9!@#$%^&*]{6,100}$/)) {
            return null;
        } else {
            return { 'invalidPassword': true };
        }
    }

    static OIDValidator(control) {
        // {6,100}           - Assert password is between 6 and 100 characters
        // (?=.*[0-9])       - Assert a string has at least one number
        if (control.value){
            if (control.value.match(/^\.[^\.].*[^\.]$/g) || control.value == "") {
                return null;
            } else {
                return { 'invalidOID': true };
            }
        }
    }

    static integerValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^\d*$/)) {
                return null;
            } else {
                return { 'invalidInteger': true };
            }
        }
    }

    static floatValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^[\+\-]?\d*\.?\d*$/g)) {
                return null;
            } else {
                return { 'invalidFloat': true };
            }
        }
    }

    static noWhiteSpaces(control) {
        if (control.value){
            if (!control.value.toString().match(/[\s]/)) {
                return null;
            } else {
                return { 'invalidWhiteSpaces': true };
            }
        }
    }
    static extraTags(control) {
        if (control.value){
            if (control.value.toString().match(/^([\w]+[=][\w]+[,]?){1,}$/g)) {
                return null;
            } else {
                return { 'invalidExtraTags': true };
            }
        }
    }


}
