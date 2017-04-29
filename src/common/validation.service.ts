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
            'invalidWhiteSpaces': 'Invalid. Can\'t contain white spaces',
            'invalidUInteger': 'Invalid Number . Must be a Unsigned (positive) Integer',
            'invalidUInteger8': 'Invalid Number . Must be a Unsigned (positive) Integer with range 0-255',
            'invalidUInteger8NotZero': 'Invalid Number . Must be a Unsigned (positive) Integer with range 1-255',
            'invalidUIntegerNotZero': 'Invalid Number . Must be a Unsigned (positive) Integer not Zero',
            'invalidUIntegerAndLessOne': 'Invalid Number . Must be a Unsigned (positive) Integer or -1'
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
        if (control.value){
            if (control.value.match(/^\.[^\.]*[^\.]$/g) ||  control.value.match(/^[\.0-9]+$/)  || control.value == "") {
                return null;
            } else {
                return { 'invalidOID': true };
            }
        }
    }

    static integerValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^(-?[1-9]+\d*)$|^0$/)) {
                return null;
            } else {
                return { 'invalidInteger': true };
            }
        }
    }

    static uintegerNotZeroValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^[1-9]+\d*$/)) {
                return null;
            } else {
                return { 'invalidUIntegerNotZero': true };
            }
        }
    }

    static uintegerValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^([1-9]+\d*)$|^0$/)) {
                return null;
            } else {
                return { 'invalidUInteger': true };
            }
        }
    }

    static uinteger8Validator(control) {
        if (control.value){
            if (control.value.toString().match(/^([1-9]+\d*)$|^0$/) && control.value < 256) {
                return null;
            } else {
                return { 'invalidUInteger8': true };
            }
        }
    }


    static uinteger8NotZeroValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^([1-9]+\d*)$/) && control.value < 256) {
                return null;
            } else {
                return { 'invalidUInteger8NotZero': true };
            }
        }
    }

    static uintegerAndLessOneValidator(control) {
        if (control.value){
            if (control.value.toString().match(/^([1-9]+\d*)$|^0$|^-1$/)) {
                return null;
            } else {
                return { 'invalidUIntegerAndLessOne': true };
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
            if (control.value.toString().match(/^([\w._:-]+[=][\w._:-]+[,]?){1,}$/g) || control.value.toString() == "") {
                return null;
            } else {
                return { 'invalidExtraTags': true };
            }
        }
    }


}
